setup_envtest_bin = $(kind_dir)/setup-envtest

# Prepare binary
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

.PHONY: local-install
local-install: export KUBECONFIG = $(KIND_KUBECONFIG)
local-install: crossplane-setup kind-load-image install-crd webhook-cert k8up-setup ## Install Operator in local cluster
	helm upgrade --install provider-postgresql charts/provider-postgresql \
		--create-namespace --namespace postgresql-system \
		--set "operator.args[0]=--log-level=2" \
		--set "operator.args[1]=operator" \
		--set podAnnotations.date="$(shell date)" \
		--values $(webhook_values) \
		--wait $(local_install_args)

.PHONY: kind-run-operator
kind-run-operator: export KUBECONFIG = $(KIND_KUBECONFIG)
kind-run-operator: kind-setup webhook-cert ## Run in Operator mode against kind cluster (you may also need `install-crd`)
	go run . -v 1 operator --webhook-tls-cert-dir .kind --operator-namespace postgresql-system


###
### Generate webhook certificates
###

.PHONY: webhook-cert
webhook-cert: $(webhook_values)

$(webhook_key):
	openssl req -x509 -newkey rsa:4096 -nodes -keyout $@ --noout -days 3650 -subj "/CN=$(webhook_service_name)" -addext "subjectAltName = DNS:$(webhook_service_name)"

$(webhook_cert): $(webhook_key)
	openssl req -x509 -key $(webhook_key) -nodes -out $@ -days 3650 -subj "/CN=$(webhook_service_name)" -addext "subjectAltName = DNS:$(webhook_service_name)"

$(webhook_values): $(webhook_cert)
	@yq -n '.webhook.caBundle="$(shell $(b64) $(webhook_cert))" | .webhook.certificate="$(shell $(b64) $(webhook_cert))" | .webhook.privateKey="$(shell $(b64) $(webhook_key))"' > $(kind_dir)/webhook-values.yaml

###
### Integration Tests
###

.PHONY: test-integration
test-integration: export ENVTEST_CRD_DIR = $(shell realpath $(envtest_crd_dir))
test-integration: $(setup_envtest_bin) .envtest_crds ## Run integration tests against code
	export KUBEBUILDER_ASSETS="$$($(setup_envtest_bin) $(ENVTEST_ADDITIONAL_FLAGS) use -i -p path '$(ENVTEST_K8S_VERSION)!')" && \
	go test -tags=integration -coverprofile cover.out -covermode atomic ./...

envtest_crd_dir ?= $(kind_dir)/crds
# Getting the version from go.mod. Doesn't work if it references a specific commit
provider_helm_version ?= $(shell go mod edit -json | jq -r '.Require[] | select(.Path == "github.com/crossplane-contrib/provider-helm") | .Version')
k8up_version ?= $(shell go mod edit -json | jq -r '.Require[] | select(.Path == "github.com/k8up-io/k8up/v2") | .Version')
provider_helm_download_root ?= https://raw.githubusercontent.com/crossplane-contrib/provider-helm/$(provider_helm_version)/package/crds
k8up_download_root ?= https://raw.githubusercontent.com/k8up-io/k8up/$(k8up_version)/config/crd/apiextensions.k8s.io/v1/base/

.envtest_crd_dir:
	@mkdir -p $(envtest_crd_dir)
	@cp -r package/crds $(kind_dir)

$(envtest_crd_dir)/helm.crossplane.io_releases.yaml: $(.envtest_crd_dir)
	curl -sSL -o $@ $(provider_helm_download_root)/helm.crossplane.io_releases.yaml

$(envtest_crd_dir)/helm.crossplane.io_providerconfigs.yaml: $(.envtest_crd_dir)
	curl -sSL -o $@ $(provider_helm_download_root)/helm.crossplane.io_providerconfigs.yaml

$(envtest_crd_dir)/k8up.io_schedules.yaml: $(.envtest_crd_dir)
	curl -sSL -o $@ $(k8up_download_root)/k8up.io_schedules.yaml

.envtest_crds: .envtest_crd_dir $(envtest_crd_dir)/helm.crossplane.io_releases.yaml $(envtest_crd_dir)/helm.crossplane.io_providerconfigs.yaml  $(envtest_crd_dir)/k8up.io_schedules.yaml

####
#### S3 Bucket
####

.PHONY: s3-credentials
s3-credentials: export KUBECONFIG = $(KIND_KUBECONFIG)
s3-credentials: minio-setup k8up-setup ## Copies Minio's S3 Bucket credentials to $(bucket_namespace)
	NEXT_WAIT_TIME=0; \
	RELEASE_EXISTS=$$(kubectl get release --no-headers --ignore-not-found=true | wc -l); \
	until [ $$NEXT_WAIT_TIME -eq 10 ] || [ $$RELEASE_EXISTS -eq 1 ]; do sleep $$((NEXT_WAIT_TIME++)); done; \
	export BUCKET=$$(yq '.buckets[0].name' test/minio-values.yaml) && \
	RELEASE_NAMESPACE=$$(kubectl get release --selector='app.kubernetes.io/instance=my-instance' -o jsonpath='{.items[*].metadata.name}') && \
	kubectl -n minio-system get secret minio-server -o yaml | \
		yq e 'del(.metadata) | .data.secretKey=.data.rootPassword | del(.data.rootPassword) | .data.accessKey=.data.rootUser | del(.data.rootUser) | .metadata.name = "s3-credentials" | .stringData.bucket=env(BUCKET) | .stringData.endpoint="http://minio-server.minio-system:9000"' | \
		kubectl -n $$RELEASE_NAMESPACE apply -f -
