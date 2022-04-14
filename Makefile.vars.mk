## These are some common variables for Make
crossplane_sentinel = $(kind_dir)/crossplane-sentinel
registry_sentinel = $(kind_dir)/registry_sentinel

PROJECT_ROOT_DIR = .
PROJECT_NAME ?= appcat-service-postgresql
PROJECT_OWNER ?= vshn

## BUILD:go
BIN_FILENAME ?= provider-postgresql

## BUILD:docker
DOCKER_CMD ?= docker

IMG_TAG ?= latest
CONTAINER_REGISTRY ?= ghcr.io
# Image URL to use all building image targets.
# NOTE: the released images are defined in .goreleaser.yml via GitHub actions.
CONTAINER_IMG ?= $(CONTAINER_REGISTRY)/$(PROJECT_OWNER)/$(PROJECT_NAME):$(IMG_TAG)
# Crossplane image reference for packaging and pushing to registry.
CROSSPLANE_REGISTRY ?= $(CONTAINER_REGISTRY)
CROSSPLANE_IMG ?= $(CROSSPLANE_REGISTRY)/$(PROJECT_OWNER)/$(PROJECT_NAME):provider-$(IMG_TAG)

## KIND:setup

# see available options in https://hub.docker.com/r/kindest/node/tags
KIND_NODE_VERSION ?= v1.23.0
KIND_IMAGE ?= docker.io/kindest/node:$(KIND_NODE_VERSION)
KIND ?= go run sigs.k8s.io/kind
KIND_KUBECONFIG ?= $(kind_dir)/kind-kubeconfig-$(KIND_NODE_VERSION)
KIND_CLUSTER ?= $(PROJECT_NAME)-$(KIND_NODE_VERSION)
