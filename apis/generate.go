//go:build generate
// +build generate

// Remove existing manifests
//go:generate rm -rf ../package/crds ../package/webhook ../charts/provider-postgresql/templates/webhook.yaml ../charts/provider-postgresql/templates/clusterrole.yaml

// Generate deepcopy methodsets and CRD manifests
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile=../.github/boilerplate.go.txt paths=./... crd:crdVersions=v1 output:artifacts:config=../package/crds

// Generate webhook manifests
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen webhook paths=./... output:artifacts:config=../charts/provider-postgresql/templates

package apis
