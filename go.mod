module github.com/monzo/vault-plugin-database-k8s-controller

go 1.13

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/fatih/structs v1.1.0
	github.com/go-test/deep v1.1.0
	github.com/hashicorp/errwrap v1.1.0
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/vault v1.13.7
	github.com/hashicorp/vault/api v1.9.2
	github.com/hashicorp/vault/sdk v0.10.0
	github.com/lib/pq v1.10.9
	github.com/mitchellh/mapstructure v1.5.0
	github.com/ory/dockertest v3.3.5+incompatible
	k8s.io/api v0.26.2
	k8s.io/apimachinery v0.26.2
	k8s.io/client-go v0.26.2
	k8s.io/klog v1.0.0
)
