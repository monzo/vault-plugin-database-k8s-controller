UPSTREAM_TAG ?= v1.2.3

build: database-k8s
	go build -o database-k8s ./database-plugin

.PHONY: rebase
rebase:
	$(eval TMPDIR := $(shell mktemp -d))
	git clone https://github.com/hashicorp/vault $(TMPDIR)
	git -C $(TMPDIR) checkout $(UPSTREAM_TAG)
	git -C $(TMPDIR) checkout -b tmp
	git -C $(TMPDIR) filter-branch -f --prune-empty --subdirectory-filter builtin/logical/database tmp
	git remote remove tmp-upstream || true
	git remote add tmp-upstream $(TMPDIR)
	git fetch tmp-upstream
	git rebase tmp-upstream/tmp
	git remote remove tmp-upstream
	rm -rf $(TMPDIR)
