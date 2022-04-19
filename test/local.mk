# Despite the registry running in Cluster, we need to load the container image with `kind load`.
# There were problems trying to pull container image from registry ("no such host") even though Crossplane could pull the package image...
.PHONY: package-install
package-install: export KUBECONFIG = $(KIND_KUBECONFIG)
package-install: CROSSPLANE_REGISTRY = localhost:5000
package-install: kind-load-image registry-setup crossplane-setup package-push ## Build and install Crossplane package in local cluster
	kubectl apply -f test/provider-postgresql.yaml
	kubectl wait --for condition=Healthy provider.pkg.crossplane.io/provider-postgresql --timeout 60s
	kubectl -n crossplane-system wait --for condition=Ready $$(kubectl -n crossplane-system get pods -o name -l pkg.crossplane.io/provider=appcat-service-postgresql) --timeout 60s

.PHONY: crossplane-setup
crossplane-setup: $(crossplane_sentinel) ## Install local Kubernetes cluster and install Crossplane

.PHONY: registry-setup
registry-setup: $(registry_sentinel) # Install docker registry in local Kubernetes cluster

$(registry_sentinel): export KUBECONFIG = $(KIND_KUBECONFIG)
$(registry_sentinel): $(KIND_KUBECONFIG)
	helm repo add twuni https://helm.twun.io
	helm upgrade --install registry twuni/docker-registry \
		--create-namespace \
		--namespace registry-system \
		--set service.type=NodePort \
		--set service.nodePort=30500 \
		--set fullnameOverride=registry \
		--wait
	@touch $@

$(crossplane_sentinel): export KUBECONFIG = $(KIND_KUBECONFIG)
$(crossplane_sentinel): $(KIND_KUBECONFIG)
	helm repo add crossplane https://charts.crossplane.io/stable
	helm upgrade --install crossplane --create-namespace --namespace crossplane-system crossplane/crossplane --set "args[0]='--debug'" --set "args[1]='--enable-composition-revisions'" --wait
	kubectl apply -f test/provider-helm.yaml
	kubectl wait --for condition=Healthy provider.pkg.crossplane.io/provider-helm --timeout 60s
	kubectl apply -f test/provider-config.yaml
	kubectl create clusterrolebinding crossplane:provider-helm-admin --clusterrole cluster-admin --serviceaccount crossplane-system:$$(kubectl get sa -n crossplane-system -o custom-columns=NAME:.metadata.name --no-headers | grep provider-helm)
	kubectl create clusterrolebinding crossplane:cluster-admin --clusterrole cluster-admin --serviceaccount crossplane-system:crossplane
	@touch $@
