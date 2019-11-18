UPSTREAM_TAG ?= v1.3.0
GOOS ?= linux
GOARCH ?= amd64

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o vault-plugin-database-k8s-controller-$(UPSTREAM_TAG)  ./database-plugin

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
