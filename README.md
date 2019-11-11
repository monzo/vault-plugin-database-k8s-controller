# vault-plugin-database-k8s-controller
A fork of Vault's database credential plugin allowing the use of an annotation on
service accounts to dynamically specify user creation statements.

Non-builtin database plugins are not supported, as sadly custom plugins cannot call out to other custom plugins.
However, all of Vault's builtin database plugins are bundled into this binary and should work as normal. The Cassandra
plugin is forked slightly, containing changes that have been upstreamed but are not yet in a stable release. This is
temporary.

Currently based on https://github.com/hashicorp/vault/tree/v1.2.3/builtin/logical/database

To rebase:
```bash
make rebase
```

## Instructions

This plugin differs from the default database plugin only if kubernetes config is provided
```bash
# if your plugin is registered as database-k8s
vault secrets enable -path=database -plugin-name=database-k8s database-k8s
vault write database/kubeconfig kubernetes_host=https://127.0.0.1 kubernetes_ca_cert=@cert jwt=@jwt
```

If this is provided, the plugin will attempt to maintain an in memory cache of all
service accounts in Kubernetes. If any service accounts contain an annotation
`monzo.com/keyspace`, the mapping from service account name to the annotation is also
stored in Vault. This is so that the mapping can be used before the cache is built from
the k8s API.

The purpose of this is to interpolate this annotation into any creation statements of a role,
to create essentially a dynamic role for every service account. If you provide a role named
like `k8s_rw_s-ledger_default` *and this role does not explicitly exist* then instead the
plugin will look up the concrete role named `rw`, and will look up the service account
`s-ledger` in the namespace `default`.

It will then replace all instances of `{{annotation}}` in the creation statements of
the concrete `rw` role with the value of the annotation on that service account.

You can also set an annotation `monzo.com/cluster` which allows you to override the db name
of the concrete `rw` role with the value of the annotation.

Annotation keys can be overridden with the `kubeconfig` endpoint, 
using `keyspace_annotation` and `db_name_annotation`.

The role names are designed such that they can support a vault policy as follows:

```hcl
path "database/creds/k8s_rw_{{identity.entity.aliases.kubernetes.metadata.service_account_name}}_{{identity.entity.aliases.kubernetes.metadata.service_account_namespace}}"
{
  capabilities = ["read"]
}
```

The role name is used for these parameters so that the plugin has the same API as its 
upstream.

## Example

```bash
$ vault write database/roles/rw \
    db_name=my-cassandra-database \
    creation_statements="CREATE USER '{{username}}' WITH PASSWORD '{{password}}' NOSUPERUSER;" \
    creation_statements="GRANT ALL PERMISSIONS ON KEYSPACE \"{{annotation}}\" TO {{username}};" \
    default_ttl="1h" \
    max_ttl="24h"
Success! Data written to: database/roles/rw

kubectl create serviceaccount s-ledger
kubectl annotate serviceaccount s-ledger monzo.com/keyspace='ledger'

$ vault read database/roles/k8s_rw_s-ledger_default
Key                      Value
---                      -----
creation_statements      [CREATE USER '{{username}}' WITH PASSWORD '{{password}}' NOSUPERUSER; GRANT ALL PERMISSIONS ON KEYSPACE "ledger" TO {{username}};]
db_name                  my-cassandra-database
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```
