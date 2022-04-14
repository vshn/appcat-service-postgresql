# This makefile is meant to be run from the parent dir

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
package_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

crossplane_bin = $(kind_dir)/kubectl-crossplane

# Build kubectl-crossplane plugin
$(crossplane_bin):
	@mkdir -p $(kind_dir)
	cd $(package_dir) && go build -o $@ github.com/crossplane/crossplane/cmd/crank

.PHONY: package
package: ## All-in-one packaging and releasing
package: package-push

.PHONY: package-provider
package-provider: controller_image = $(shell yq e '.spec.controller.image' $(package_dir)/crossplane.yaml)
package-provider: $(crossplane_bin) ## Build Crossplane package
	@rm -rf package/*.xpkg
	@yq -i e '.spec.controller.image="$(CONTAINER_IMG)"' $(package_dir)/crossplane.yaml
	$(crossplane_bin) build provider -f $(package_dir)
	@yq -i e '.spec.controller.image="$(controller_image)"' $(package_dir)/crossplane.yaml
	@ls $(package_dir)/*.xpkg

.PHONY: package-push
package-push: pkg_file = $(shell ls $(package_dir)/*.xpkg)
package-push: package-provider ## Push Crossplane package to container registry
	$(crossplane_bin) push provider -f $(pkg_file) $(CROSSPLANE_IMG)
