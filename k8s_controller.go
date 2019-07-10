package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/logical"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"path"
	"regexp"
	"time"
)

var annotationKey = "monzo.com/keyspace"

var saCache cache.Store

// watchServiceAccounts is called on plugin start and attempts to maintain an
// in-memory cache of all service accounts.
func watchServiceAccounts(kubeconfig string) error {
	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return err
	}

	client, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	lw := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "serviceaccounts", "", fields.Everything())

	saCache = cache.NewStore(keyFunc)

	reflector := cache.NewReflector(lw, &v1.ServiceAccount{}, saCache, time.Hour)

	stopCh := make(chan struct{})
	go reflector.Run(stopCh)

	return nil
}

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

func getAnnotationForObj(obj interface{}) (string, error) {
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

func getServiceAccountAnnotation(ctx context.Context, s logical.Storage, namespace, svcAccountName string) (string, error) {
	// first try from the cache
	sa, exists, err := saCache.GetByKey(path.Join(namespace, svcAccountName))
	if err != nil {
		return "", err
	}

	if exists {
		return getAnnotationForObj(sa)
	}

	// now try from durable storage
	key := path.Join("config", "serviceaccount", namespace, svcAccountName)
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
func syncServiceAccounts(ctx context.Context, req *logical.Request) error {
	if saCache == nil {
		return nil
	}

	sas := saCache.List()

	klog.Infof("Syncing %d service accounts", len(sas))
	for _, sa := range sas {
		annotation, err := getAnnotationForObj(sa)
		if err != nil {
			klog.Errorf("error getting annotation for object: %s", err)
			continue
		}

		if annotation == "" {
			continue
		}

		key, err := keyFunc(sa)
		if err != nil {
			return err
		}

		// store in config/serviceaccount/default/s-ledger
		entry, err := logical.StorageEntryJSON(path.Join("config", "serviceaccount", key), annotation)
		if err != nil {
			return err
		}

		err = req.Storage.Put(ctx, entry)
		if err != nil {
			return err
		}

		klog.Infof("wrote to path %s", entry.Key)
	}

	return nil
}
