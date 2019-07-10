# vault-plugin-database-k8s-controller
A fork of Vault's database credential plugin allowing the use of an annotation on
service accounts to dynamically specify user creation statements.

Currently based on https://github.com/hashicorp/vault/tree/v1.1.3/builtin/logical/database

To rebase:
```bash
cd ~/src/github.com/hashicorp/vault
git checkout $tag
git checkout -b tmp
git filter-branch -f --prune-empty --subdirectory-filter builtin/logical/database tmp
cd ~/src/github.com/monzo/vault-plugin-database-k8s-controller 
git remote add upstream ~/src/github.com/hashicorp/vault
git fetch upstream
git rebase upstream/tmp
```

## Instructions

This plugin differs from the default database plugin only if a kubeconfig option is provided on mount:
```bash
# if your plugin is registered as database-k8s
vault secrets enable -path=database -plugin-name=database-k8s -options=kubeconfig=$configpath database-k8s
```

If this is provided, the plugin will attempt to maintain an in memory cache of all
service accounts in Kubernetes. If any service accounts contain an annotation
`monzo.com/keyspace`, the mapping from service account name to the annotation is also
stored in Vault. This is so that the mapping can be used before the cache is built from
the k8s API.

A custom annotation key can be provided on `secrets enable` with the option `annotation_key`

The purpose of this is to interpolate this annotation into any creation statements of a role,
to create essentially a dynamic role for every service account. If you provide a role named
like `k8s_rw_default_s-ledger` *and this role does not explicitly exist* then instead the
plugin will look up the concrete role named `rw`, and will look up the service account
`s-ledger` in the namespace `default`.

It will then replace all instances of `{{annotation}}` in the creation statements of
the concrete `rw` role with the value of the annotation on that service account.

The role names are designed such that they can support a vault policy as follows:

```hcl
path "database/creds/k8s_rw_{{identity.entity.aliases.kubernetes.metadata.service_account_namespace}}_{{identity.entity.aliases.kubernetes.metadata.service_account_name}}"
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
    creation_statements="GRANT SELECT ON KEYSPACE \"{{annotation}}\" TO {{username}};" \
    default_ttl="1h" \
    max_ttl="24h"
Success! Data written to: database/roles/rw

kubectl create serviceaccount s-ledger
kubectl annotate serviceaccount s-ledger monzo.com/keyspace='ledger'

$ vault read database/roles/k8s_rw_default_s-ledger
Key                      Value
---                      -----
creation_statements      [CREATE USER '{{username}}' WITH PASSWORD '{{password}}' NOSUPERUSER; GRANT SELECT ON KEYSPACE "ledger" TO {{username}};]
db_name                  my-cassandra-database
default_ttl              1h
max_ttl                  24h
renew_statements         []
revocation_statements    []
rollback_statements      []
```
