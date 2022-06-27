package pipe

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clientKey struct{}

func SetClientInContext(ctx context.Context, c client.Client) {
	pipeline.StoreInContext(ctx, clientKey{}, c)
}
func GetClientFromContext(ctx context.Context) client.Client {
	return pipeline.MustLoadFromContext(ctx, clientKey{}).(client.Client)
}
