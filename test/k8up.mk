k8up_sentinel = $(kind_dir)/k8up_sentinel
minio_sentinel = $(kind_dir)/minio_sentinel

.PHONY: k8up-setup
k8up-setup: minio-setup $(k8up_sentinel) ## Install K8up operator

$(k8up_sentinel): export KUBECONFIG = $(KIND_KUBECONFIG)
$(k8up_sentinel): $(KIND_KUBECONFIG)
	helm repo add appuio https://charts.appuio.ch
	kubectl apply -f https://github.com/k8up-io/k8up/releases/latest/download/k8up-crd.yaml
	helm upgrade --install k8up appuio/k8up \
		--create-namespace --namespace k8up-system \
		--wait
	kubectl -n k8up-system wait --for condition=Available deployment/k8up --timeout 60s
	@touch $@

.PHONY: minio-setup
minio-setup: $(minio_sentinel) ## Install Minio S3

$(minio_sentinel): export KUBECONFIG = $(KIND_KUBECONFIG)
$(minio_sentinel): $(KIND_KUBECONFIG)
	helm repo add minio https://charts.min.io
	helm upgrade --install minio minio/minio \
		--create-namespace --namespace minio-system \
		--values test/minio-values.yaml \
		--wait
	@touch $@
