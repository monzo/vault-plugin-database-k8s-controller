package database

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// watchServiceAccounts is called on plugin start and attempts to maintain an
// in-memory cache of all service accounts.
func (b *databaseBackend) watchServiceAccounts(kubeconfig *kubeConfig) (func(), error) {
	b.logger.Info("kubeconfig provided; will watch for Kubernetes service accounts")

	config := &rest.Config{
		Host:        kubeconfig.Host,
		BearerToken: kubeconfig.JWT,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(kubeconfig.CACert),
		},
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	lw := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "serviceaccounts", "", fields.Everything())

	reflector := cache.NewReflector(lw, &v1.ServiceAccount{}, b.saCache, time.Hour)

	stopCh := make(chan struct{})
	go reflector.Run(stopCh)

	return func() {
		b.logger.Info("Closing reflector")
		close(stopCh)
	}, nil
}

// keyFunc is very similar to cache.MetaNamespaceKeyFunc except when
// there's no namespace specified it uses "default"
func keyFunc(obj interface{}) (string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("object has no meta: %v", err)
	}
	if len(meta.GetNamespace()) > 0 {
		return meta.GetNamespace() + "/" + meta.GetName(), nil
	}
	return "default/" + meta.GetName(), nil
}

const nameRegexStr = `^[\w.]+$`

var nameRegex = regexp.MustCompile(nameRegexStr)

// getAnnotationForObj pulls the configured annotation key out of a k8s object,
// checking it against a fairly restrictive regex to avoid injection
func (b *databaseBackend) getAnnotationForObj(annotationKey string, obj interface{}) (string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return "", err
	}

	annotations := meta.GetAnnotations()

	if annotations == nil {
		return "", nil
	}

	if value, ok := annotations[annotationKey]; ok {
		if len(value) > 0 {
			if !nameRegex.MatchString(value) {
				return "", errors.New(fmt.Sprintf("annotation %s did not match regex %s", value, nameRegexStr))
			}

			return value, nil
		}
	}

	return "", nil
}

// getServiceAccountAnnotation tries two strategies to find the annotation value for a service account.
// First it tries to read the service account out of the reflector cache. However this may not be populated
// if the plugin just started. If not found there, it reads Vault storage in case the plugin has ever synced
// this service account before and stored it persistently.
func (b *databaseBackend) getServiceAccountAnnotation(ctx context.Context, s logical.Storage, namespace, svcAccountName string) (string, error) {
	// first try from the cache
	sa, exists, err := b.saCache.GetByKey(path.Join(namespace, svcAccountName))
	if err != nil {
		return "", err
	}

	if exists {
		config, err := b.kubeconfig(ctx, s)
		if err != nil {
			return "", err
		}

		if config != nil {
			return b.getAnnotationForObj(config.AnnotationKey, sa)
		}
	}

	// now try from durable storage
	key := path.Join("serviceaccount", namespace, svcAccountName)
	entry, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if entry == nil {
		return "", nil
	}

	var value string

	if err := entry.DecodeJSON(&value); err != nil {
		return "", err
	}

	return value, nil
}

// syncServiceAccounts lists all known service accounts to obtain a mapping of name to annotation
// and stores this mapping durably in Vault. This allows us to load it immediately on plugin start.
// Vault should call this function every minute.
func (b *databaseBackend) syncServiceAccounts(ctx context.Context, req *logical.Request) error {
	sas := b.saCache.List()

	if len(sas) == 0 {
		return nil
	}

	config, err := b.kubeconfig(ctx, req.Storage)
	if err != nil {
		return err
	}

	b.logger.Debug(fmt.Sprintf("Syncing %d service accounts", len(sas)))

	written := map[string]struct{}{}
	for _, sa := range sas {
		annotation, err := b.getAnnotationForObj(config.AnnotationKey, sa)
		if err != nil {
			b.logger.Error("error getting annotation for object: %s", err)
			continue
		}

		if annotation == "" {
			continue
		}

		key, err := keyFunc(sa)
		if err != nil {
			return err
		}

		// store in serviceaccount/default/s-ledger
		entry, err := logical.StorageEntryJSON(path.Join("serviceaccount", key), annotation)
		if err != nil {
			return err
		}

		err = req.Storage.Put(ctx, entry)
		if err != nil {
			return err
		}

		written[entry.Key] = struct{}{}
	}

	// we should also delete any service accounts that no longer have the annotation
	keys, err := logical.CollectKeysWithPrefix(ctx, req.Storage, "serviceaccount/")
	if err != nil {
		return err
	}

	var deleted int
	for _, k := range keys {
		if _, ok := written[k]; !ok {
			if err := req.Storage.Delete(ctx, k); err != nil {
				return err
			}
			deleted++
		}
	}

	b.logger.Debug(fmt.Sprintf("wrote %d service accounts to storage, deleted %d", len(written), deleted))

	return nil
}
