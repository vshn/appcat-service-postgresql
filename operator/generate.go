//go:build generate
// +build generate

// Generate webhook manifests
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen rbac:roleName=manager-role paths=./... output:artifacts:config=../chart/templates

package operator
