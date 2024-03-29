# Set Shell to bash, otherwise some targets fail with dash/zsh etc.
SHELL := /bin/bash

# Disable built-in rules
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
.SECONDARY:
.DEFAULT_GOAL := help

# General variables
include Makefile.vars.mk

# Following includes do not print warnings or error if files aren't found
# Optional Documentation module.
-include docs/antora-preview.mk docs/antora-build.mk
# Optional kind module
-include kind/kind.mk
# Chart-related
-include charts/charts.mk
# Local Env & testing
-include test/local.mk test/crossplane.mk test/k8up.mk

.PHONY: help
help: ## Show this help
	@grep -E -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: build-bin build-docker ## All-in-one build

.PHONY: build-bin
build-bin: export CGO_ENABLED = 0
build-bin: fmt vet ## Build binary
	@go build -o $(BIN_FILENAME) .

.PHONY: build-docker
build-docker: build-bin ## Build docker image
	$(DOCKER_CMD) build -t $(CONTAINER_IMG) .

.PHONY: test
test: test-go ## All-in-one test

.PHONY: test-go
test-go: ## Run unit tests against code
	go test -race -coverprofile cover.out -covermode atomic ./...

.PHONY: fmt
fmt: ## Run 'go fmt' against code
	go fmt ./...

.PHONY: vet
vet: ## Run 'go vet' against code
	go vet ./...

.PHONY: lint
lint: fmt vet generate ## All-in-one linting
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code

.PHONY: generate
generate: generate-go generate-docs ## All-in-one code generation

.PHONY: generate-go
generate-go: ## Generate Go artifacts
	@go generate ./...

.PHONY: generate-docs
generate-docs: generate-go ## Generate example code snippets for documentation
	@yq e 'del(.metadata.creationTimestamp) | del(.metadata.generation) | del(.status)' package/samples/postgresql.appcat.vshn.io_postgresqlstandalone.yaml > $(docs_moduleroot_dir)/examples/standalone.yaml
	@yq e '.spec.package="crossplane/provider-helm:$(provider_helm_version)"' test/provider-helm.yaml > $(docs_moduleroot_dir)/examples/installation/provider-helm.yaml
	@yq e 'del(.spec.helmReleases) | del(.metadata.creationTimestamp) | del(.status)' package/samples/postgresql.appcat.vshn.io_postgresqlstandaloneoperatorconfig.yaml > $(docs_moduleroot_dir)/examples/installation/config-v14.yaml
	@yq e 'del(.metadata.creationTimestamp) | del(.status)' package/samples/helm.crossplane.io_providerconfig.yaml > $(docs_moduleroot_dir)/examples/installation/providerconfig-helm.yaml
	@cp test/rbac.yaml test/controller-config.yaml $(docs_moduleroot_dir)/examples/installation/

.PHONY: install-crd
install-crd: export KUBECONFIG = $(KIND_KUBECONFIG)
install-crd: generate kind-setup ## Install CRDs into cluster
	kubectl apply -f package/crds

.PHONY: install-samples
install-samples: export KUBECONFIG = $(KIND_KUBECONFIG)
install-samples: generate-go install-crd ## Install samples into cluster
	yq package/samples/*.yaml | kubectl apply -f -

.PHONY: delete-instance
delete-instance: export KUBECONFIG = $(KIND_KUBECONFIG)
delete-instance:  ## Deletes sample instance if it exists
	kubectl delete -f package/samples/postgresql.appcat.vshn.io_postgresqlstandalone.yaml --ignore-not-found=true

.PHONY: run-operator
run-operator: ## Run in Operator mode against your current kube context
	go run . -v 1 operator

.PHONY: clean
clean: kind-clean ## Cleans local build artifacts
	rm -rf docs/node_modules $(docs_out_dir) dist .cache package/*.xpkg
	$(DOCKER_CMD) rmi $(CONTAINER_IMG) || true

.PHONY: release-prepare
release-prepare: generate-go ## Prepares artifacts for releases
	@cat package/crds/*.yaml | yq > .github/crds.yaml
