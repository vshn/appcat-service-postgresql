.PHONY: chart-prepare
chart-prepare: ## Prepare the Helm charts
	@find charts -type f -name Makefile | sed 's|/[^/]*$$||' | xargs -I '%' make -C '%' prepare
