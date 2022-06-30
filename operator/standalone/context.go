package standalone

import (
	"context"
	"fmt"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	corev1 "k8s.io/api/core/v1"
	"reflect"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clientKey struct{}
type instanceKey struct{}
type configKey struct{}
type helmReleaseKey struct{}
type helmValuesKey struct{}
type deploymentNamespaceKey struct{}

func setClientInContext(ctx context.Context, c client.Client) {
	pipeline.StoreInContext(ctx, clientKey{}, c)
}
func getClientFromContext(ctx context.Context) client.Client {
	return getFromContextOrPanic(ctx, clientKey{}).(client.Client)
}

func setInstanceInContext(ctx context.Context, obj *v1alpha1.PostgresqlStandalone) {
	pipeline.StoreInContext(ctx, instanceKey{}, obj)
}

func getInstanceFromContext(ctx context.Context) *v1alpha1.PostgresqlStandalone {
	return getFromContextOrPanic(ctx, instanceKey{}).(*v1alpha1.PostgresqlStandalone)
}

func setConfigInContext(ctx context.Context, config *v1alpha1.PostgresqlStandaloneOperatorConfig) {
	pipeline.StoreInContext(ctx, configKey{}, config)
}

func getConfigFromContext(ctx context.Context) *v1alpha1.PostgresqlStandaloneOperatorConfig {
	return getFromContextOrPanic(ctx, configKey{}).(*v1alpha1.PostgresqlStandaloneOperatorConfig)
}

func setHelmReleaseInContext(ctx context.Context, helmRelease *helmv1beta1.Release) {
	pipeline.StoreInContext(ctx, helmReleaseKey{}, helmRelease)
}

func getHelmReleaseFromContext(ctx context.Context) *helmv1beta1.Release {
	return getFromContextOrPanic(ctx, helmReleaseKey{}).(*helmv1beta1.Release)
}

func setHelmValuesInContext(ctx context.Context, helmValues helmvalues.V) {
	pipeline.StoreInContext(ctx, helmValuesKey{}, helmValues)
}

func getHelmValuesFromContext(ctx context.Context) helmvalues.V {
	return getFromContextOrPanic(ctx, helmValuesKey{}).(helmvalues.V)
}

func setDeploymentNamespaceInContext(ctx context.Context, namespace *corev1.Namespace) {
	pipeline.StoreInContext(ctx, deploymentNamespaceKey{}, namespace)
}

func getDeploymentNamespaceFromContext(ctx context.Context) *corev1.Namespace {
	return getFromContextOrPanic(ctx, deploymentNamespaceKey{}).(*corev1.Namespace)
}

func setHelmChartInContext(ctx context.Context, namespace *v1alpha1.ChartMeta) {
	pipeline.StoreInContext(ctx, v1alpha1.ChartMeta{}, namespace)
}

func getHelmChartFromContext(ctx context.Context) *v1alpha1.ChartMeta {
	return getFromContextOrPanic(ctx, v1alpha1.ChartMeta{}).(*v1alpha1.ChartMeta)
}

func getFromContextOrPanic(ctx context.Context, key any) any {
	val, exists := pipeline.LoadFromContext(ctx, key)
	if !exists {
		keyName := reflect.TypeOf(key).Name()
		panic(fmt.Errorf("key %q does not exist in the given context", keyName))
	}
	return val
}
