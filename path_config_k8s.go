package database

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const kubeconfigPath string = "kubeconfig"

// pathKubeconfig returns configuration for Kubernetes
func pathKubeconfig(b *databaseBackend) []*framework.Path {
	return []*framework.Path{{
		Pattern: "kubeconfig$",
		Fields: map[string]*framework.FieldSchema{
			"kubernetes_host": {
				Type:        framework.TypeString,
				Description: "Host must be a host string, a host:port pair, or a URL to the base of the Kubernetes API server.",
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Kubernetes Host",
				},
				Required: true,
			},

			"kubernetes_ca_cert": {
				Type:        framework.TypeString,
				Description: "PEM encoded CA cert for use by the TLS client used to talk with the API.",
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Kubernetes CA Certificate",
				},
			},
			"jwt": {
				Type:        framework.TypeString,
				Description: "A JWT used to access the K8S API to read service accounts.",
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "JWT",
				},
				Required: true,
			},
			"keyspace_annotation": {
				Type:        framework.TypeString,
				Description: "Keyspace annotation to look for in service accounts, where the value is interpolated into the base role.",
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Keyspace Annotation",
				},
				Default: "monzo.com/keyspace",
			},
			"db_name_annotation": {
				Type:        framework.TypeString,
				Description: "Database name to look for in service accounts, where the value (if set) overrides the value in the base role.",
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Database Name Annotation",
				},
				Default: "monzo.com/cluster",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathKubeconfigWrite(),
			logical.CreateOperation: b.pathKubeconfigWrite(),
			logical.ReadOperation:   b.pathKubeconfigRead(),
		},

		HelpSynopsis:    confHelpSyn,
		HelpDescription: confHelpDesc,
	}}
}

// kubeconfig takes a storage object and returns a kubeConfig object
func (b *databaseBackend) kubeconfig(ctx context.Context, s logical.Storage) (*kubeConfig, error) {
	raw, err := s.Get(ctx, kubeconfigPath)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	conf := &kubeConfig{}
	if err := json.Unmarshal(raw.Value, conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// pathConfigWrite handles create and update commands to the config
func (b *databaseBackend) pathKubeconfigRead() framework.OperationFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
		if config, err := b.kubeconfig(ctx, req.Storage); err != nil {
			return nil, err
		} else if config == nil {
			return nil, nil
		} else {
			// Create a map of data to be returned
			resp := &logical.Response{
				Data: map[string]interface{}{
					"kubernetes_host":     config.Host,
					"kubernetes_ca_cert":  config.CACert,
					"keyspace_annotation": config.KeyspaceAnnotation,
					"db_name_annotation":  config.DBNameAnnotation,
				},
			}

			return resp, nil
		}
	}
}

// pathConfigWrite handles create and update commands to the config
func (b *databaseBackend) pathKubeconfigWrite() framework.OperationFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
		host := data.Get("kubernetes_host").(string)
		if host == "" {
			return logical.ErrorResponse("no host provided"), nil
		}

		caCert := data.Get("kubernetes_ca_cert").(string)
		if len(caCert) == 0 {
			return logical.ErrorResponse("kubernetes_ca_cert must be set"), nil
		}

		jwt := data.Get("jwt").(string)
		if jwt == "" {
			return logical.ErrorResponse("jwt must be set"), nil
		}
		keyspaceAnnotationKey := data.Get("keyspace_annotation").(string)
		dbNameAnnotationKey := data.Get("db_name_annotation").(string)
		config := &kubeConfig{
			Host:               host,
			CACert:             caCert,
			JWT:                jwt,
			KeyspaceAnnotation: keyspaceAnnotationKey,
			DBNameAnnotation:   dbNameAnnotationKey,
		}

		entry, err := logical.StorageEntryJSON(kubeconfigPath, config)
		if err != nil {
			return nil, err
		}

		if err := req.Storage.Put(ctx, entry); err != nil {
			return nil, err
		}

		b.stopMtx.Lock()
		defer b.stopMtx.Unlock()

		if b.stopWatch != nil {
			b.stopWatch()
		}

		stop, err := b.watchServiceAccounts(config)
		if err != nil {
			return nil, err
		}
		b.stopWatch = stop

		return nil, nil
	}
}

// kubeConfig contains the public key certificate used to verify the signature
// on the service account JWTs
type kubeConfig struct {
	// Host is the url string for the kubernetes API
	Host string `json:"host"`
	// CACert is the CA Cert to use to call into the kubernetes API
	CACert string `json:"ca_cert"`
	// JWT is the bearer to use during the API call
	JWT string `json:"jwt"`
	// KeyspaceAnnotation is the annotation key to look for in service accounts to interpolate into statements
	KeyspaceAnnotation string `json:"keyspace_annotation"`
	// DBNameAnnotation is the annotation key to look for in service accounts to override database name for a role
	DBNameAnnotation string `json:"db_name_annotation"`
}

const confHelpSyn = `Configures the JWT Public Key and Kubernetes API information.`
const confHelpDesc = `
The k8s-controller database reads service account objects via the k8s API.
This endpoint configures the necessary information to access the Kubernetes API.
`
