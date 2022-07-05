package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchServiceFn returns a function that gets the service object and puts it into the context.
func FetchServiceFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

		service := &corev1.Service{}
		err := kube.Get(ctx, client.ObjectKey{Name: "postgresql", Namespace: instance.Status.HelmChart.DeploymentNamespace}, service)
		pipeline.StoreInContext(ctx, ServiceKey{}, service)
		return err
	}
}
