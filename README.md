# vault-plugin-database-k8s-controller
A fork of Vault's database credential plugin allowing the use of annotations on service accounts as parameters in statements

Currently based on https://github.com/hashicorp/vault/tree/v1.2.3/builtin/logical/database

To rebase:
```
cd ~/src/github.com/hashicorp/vault
git checkout $tag
git checkout -b tmp
git filter-branch -f --prune-empty --subdirectory-filter builtin/logical/database tmp
cd ~/src/github.com/monzo/vault-plugin-database-k8s-controller 
git remote add upstream ~/src/github.com/hashicorp/vault
git fetch upstream
git rebase upstream/tmp
``
