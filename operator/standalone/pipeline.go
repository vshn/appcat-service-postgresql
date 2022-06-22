package standalone

import (
	"context"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clientKey struct{}

func setClientInContext(ctx context.Context, c client.Client) {
	pipeline.StoreInContext(ctx, clientKey{}, c)
}
func getClientFromContext(ctx context.Context) client.Client {
	return pipeline.MustLoadFromContext(ctx, clientKey{}).(client.Client)
}

type instanceKey struct{}

func setInstanceInContext(ctx context.Context, obj *v1alpha1.PostgresqlStandalone) {
	pipeline.StoreInContext(ctx, instanceKey{}, obj)
}
func getInstanceFromContext(ctx context.Context) *v1alpha1.PostgresqlStandalone {
	return pipeline.MustLoadFromContext(ctx, instanceKey{}).(*v1alpha1.PostgresqlStandalone)
}

type connectionSecretKey struct{}

func getObjectFromContext[V client.Object](ctx context.Context, key any, into V) V {
	obj := pipeline.MustLoadFromContext(ctx, key)
	into = obj.(V)
	return into
}

func setObjectInContext(ctx context.Context, key any, obj client.Object) {
	pipeline.StoreInContext(ctx, key, obj)
}
