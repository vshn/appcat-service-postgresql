package standalone

import (
	"context"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
)

type instanceKey struct{}

func setInstanceInContext(ctx context.Context, obj *v1alpha1.PostgresqlStandalone) {
	pipeline.StoreInContext(ctx, instanceKey{}, obj)
}
func getInstanceFromContext(ctx context.Context) *v1alpha1.PostgresqlStandalone {
	return pipeline.MustLoadFromContext(ctx, instanceKey{}).(*v1alpha1.PostgresqlStandalone)
}
