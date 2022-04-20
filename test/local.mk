# Despite the registry running in Cluster, we need to load the container image with `kind load`.
# There were problems trying to pull container image from registry ("no such host") even though Crossplane could pull the package image...
.PHONY: package-install
package-install: export KUBECONFIG = $(KIND_KUBECONFIG)
package-install: kind-load-image install-crd ## Install Operator in local cluster
	helm upgrade --install provider-postgresql chart \
		--create-namespace --namespace crossplane-system \
		--set "args[0]='--log-level=2" \
		--set "args[1]='operator'" \
		--set podAnnotations.date="$(shell date)" \
		--wait

.PHONY: kind-run-operator
kind-run-operator: export KUBECONFIG = $(KIND_KUBECONFIG)
kind-run-operator: kind-setup ## Run in Operator mode against kind cluster (you may also need `install-crd`)
	go run . -v 1 operator
