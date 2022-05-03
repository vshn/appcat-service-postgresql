setup_envtest_bin = $(kind_dir)/setup-envtest

# Prepare kind binary
# We need to set the Go arch since the binary is meant for the user's OS.
$(setup_envtest_bin): export GOOS = $(shell go env GOOS)
$(setup_envtest_bin): export GOARCH = $(shell go env GOARCH)
$(setup_envtest_bin):
	@mkdir -p $(kind_dir)
	cd test && go build -o $@ sigs.k8s.io/controller-runtime/tools/setup-envtest
	$@ $(ENVTEST_ADDITIONAL_FLAGS) use '$(ENVTEST_K8S_VERSION)!'
	chmod -R +w $(kind_dir)/k8s

webhook_key = $(kind_dir)/tls.key
webhook_cert = $(kind_dir)/tls.crt
webhook_service_name = provider-postgresql.postgresql-system.svc
webhook_values = $(kind_dir)/webhook-values.yaml

ifeq ($(shell uname -s),Darwin)
	b64 := base64
else
  b64 := base64 -w0
endif

# Despite the registry running in Cluster, we need to load the container image with `kind load`.
.PHONY: local-install
local-install: export KUBECONFIG = $(KIND_KUBECONFIG)
local-install: kind-load-image install-crd webhook-cert ## Install Operator in local cluster
	helm upgrade --install provider-postgresql chart \
		--create-namespace --namespace postgresql-system \
		--set "args[0]='--log-level=2" \
		--set "args[1]='operator'" \
		--set podAnnotations.date="$(shell date)" \
		--values $(webhook_values) \
		--wait

.PHONY: kind-run-operator
kind-run-operator: export KUBECONFIG = $(KIND_KUBECONFIG)
kind-run-operator: kind-setup webhook-cert ## Run in Operator mode against kind cluster (you may also need `install-crd`)
	go run . -v 1 operator --webhook-tls-cert-dir .kind

.PHONY: webhook-cert
webhook-cert: $(webhook_values)

$(webhook_key):
	openssl req -x509 -newkey rsa:4096 -nodes -keyout $@ --noout -days 3650 -subj "/CN=$(webhook_service_name)" -addext "subjectAltName = DNS:$(webhook_service_name)"

$(webhook_cert): $(webhook_key)
	openssl req -x509 -key $(webhook_key) -nodes -out $@ -days 3650 -subj "/CN=$(webhook_service_name)" -addext "subjectAltName = DNS:$(webhook_service_name)"

$(webhook_values): $(webhook_cert)
	@yq -n '.webhook.caBundle="$(shell $(b64) $(webhook_cert))" | .webhook.certificate="$(shell $(b64) $(webhook_cert))" | .webhook.privateKey="$(shell $(b64) $(webhook_key))"' > $(kind_dir)/webhook-values.yaml

.PHONY: test-integration
test-integration: export ENVTEST_CRD_DIR = $(shell realpath $(PROJECT_ROOT_DIR))/package/crds
test-integration: $(setup_envtest_bin) ## Run integration tests against code
	export KUBEBUILDER_ASSETS="$$($(setup_envtest_bin) $(ENVTEST_ADDITIONAL_FLAGS) use -i -p path '$(ENVTEST_K8S_VERSION)!')" && \
	go test -tags=integration -coverprofile cover.out -covermode atomic ./...
