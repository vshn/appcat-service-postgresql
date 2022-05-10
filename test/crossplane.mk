test_dir ?= test
crossplane_sentinel = $(kind_dir)/crossplane_sentinel
provider_helm_sentinel = $(kind_dir)/provider_helm_sentinel

.PHONY: crossplane-setup
crossplane-setup: $(crossplane_sentinel) $(provider_helm_sentinel) ## Install local Kubernetes cluster and install Crossplane

$(crossplane_sentinel): export KUBECONFIG = $(KIND_KUBECONFIG)
$(crossplane_sentinel): $(KIND_KUBECONFIG)
	helm repo add crossplane https://charts.crossplane.io/stable
	helm upgrade --install crossplane crossplane/crossplane \
		--create-namespace --namespace crossplane-system \
		--set "args[0]='--debug'" \
		--set "args[1]='--enable-composition-revisions'" \
		--set webhooks.enabled=true \
		--wait
	kubectl apply -f $(test_dir)/clusterrolebinding-crossplane.yaml
	@touch $@

$(provider_helm_sentinel): export KUBECONFIG = $(KIND_KUBECONFIG)
$(provider_helm_sentinel): $(crossplane_sentinel)
	yq e '.spec.package="crossplane/provider-helm:$(provider_helm_version)"' $(test_dir)/provider-helm.yaml | kubectl apply -f -
	kubectl wait --for condition=Healthy provider.pkg.crossplane.io/provider-helm --timeout 60s
	yq e '.subjects[0].name="'$$(kubectl get sa -n crossplane-system -o yaml | yq e '.items[].metadata.name | select(match("provider-helm"))')'"' $(test_dir)/clusterrolebinding-provider-helm.yaml | kubectl apply -f -
	@touch $@
