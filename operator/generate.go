//go:build generate
// +build generate

// Generate manifests
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen rbac:roleName=manager-role paths=./... output:artifacts:config=../package/rbac

package operator
