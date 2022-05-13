helm_docs_bin := $(kind_dir)/helm-docs

# Prepare binary
# We need to set the Go arch since the binary is meant for the user's OS.
$(helm_docs_bin): export GOOS = $(shell go env GOOS)
$(helm_docs_bin): export GOARCH = $(shell go env GOARCH)
$(helm_docs_bin):
	@mkdir -p $(kind_dir)
	cd charts && go build -o $@ github.com/norwoodj/helm-docs/cmd/helm-docs

.PHONY: chart-prepare
chart-prepare: generate-go ## Prepare the Helm charts
	@mkdir -p charts/.artifacts
	@find charts -type f -name Makefile | sed 's|/[^/]*$$||' | xargs -I '%' make -C '%' clean prepare

.PHONY: chart-docs
chart-docs: $(helm_docs_bin) ## Creates the Chart READMEs from template and values.yaml files
	@$(helm_docs_bin) \
		--template-files ./.github/helm-docs-header.gotmpl.md \
		--template-files README.gotmpl.md \
		--template-files ./.github/helm-docs-footer.gotmpl.md

.PHONY: chart-lint
chart-lint: chart-prepare chart-docs ## Lint charts
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code
