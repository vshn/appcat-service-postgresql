helm_docs_bin := $(kind_dir)/helm-docs

# Prepare binary
# We need to set the Go arch since the binary is meant for the user's OS.
$(helm_docs_bin): export GOOS = $(shell go env GOOS)
$(helm_docs_bin): export GOARCH = $(shell go env GOARCH)
$(helm_docs_bin):
	@mkdir -p $(kind_dir)
	cd charts && go build -o $@ github.com/norwoodj/helm-docs/cmd/helm-docs

.PHONY: chart-prepare
chart-prepare: ## Prepare the Helm charts
	@find charts -type f -name Makefile | sed 's|/[^/]*$$||' | xargs -I '%' make -C '%' prepare

.PHONY: chart-docs
chart-docs: $(helm_docs_bin) ## Creates the Chart READMEs from template and values.yaml files
	@$(helm_docs_bin) \
		--template-files ./.github/helm-docs-header.gotmpl.md \
		--template-files README.gotmpl.md \
		--template-files ./.github/helm-docs-footer.gotmpl.md

.PHONY: chart-lint
chart-lint: export charts_dir = charts
chart-lint: ## Checks if chart versions have been changed
	@echo "    If this target fails, one of the listed charts below has not its version updated!"
	@changed_charts=$$(git diff --dirstat=files,0 origin/master..HEAD -- $(charts_dir) | cut -d '/' -f 2 | uniq) ; \
	  echo $$changed_charts ; echo ;  \
	  for dir in $$changed_charts; do git diff origin/master..HEAD -- "$(charts_dir)/$${dir}/Chart.yaml" | grep -H --label=$${dir} "+version"; done
