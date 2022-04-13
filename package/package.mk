mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
package_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

crossplane_bin = $(kind_dir)/kubectl-crossplane

$(crossplane_bin): $(kind_dir) ## Build kubectl-crossplane plugin
	cd $(package_dir) && go build -o $@ github.com/crossplane/crossplane/cmd/crank

.PHONY: package
package: ## All-in-one packaging
package: package-provider

.PHONY: package-provider
package-provider: $(crossplane_bin) ## Build Crossplane package
	@rm -rf package/*.xpkg
	$(crossplane_bin) build provider -f $(package_dir)
