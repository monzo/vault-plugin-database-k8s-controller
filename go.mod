module github.com/monzo/vault-plugin-database-k8s-controller

go 1.13

require (
	github.com/fatih/structs v1.1.0
	github.com/go-test/deep v1.0.2
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-uuid v1.0.2-0.20191001231223-f32f5fe8d6a8
	github.com/hashicorp/vault v1.3.0
	github.com/hashicorp/vault/api v1.0.5-0.20191108163347-bdd38fca2cff
	github.com/hashicorp/vault/sdk v0.1.14-0.20191112033314-390e96e22eb2
	github.com/lib/pq v1.2.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/ory/dockertest v3.3.4+incompatible
	k8s.io/api v0.17.16
	k8s.io/apimachinery v0.17.16
	k8s.io/client-go v0.17.16
	k8s.io/klog v1.0.0
)
