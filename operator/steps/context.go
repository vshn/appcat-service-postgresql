package steps

import (
	"context"
	"fmt"
	"reflect"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClientKey identifies the Kubernetes client in the context.
type ClientKey struct{}

// InstanceKey identifies the v1alpha1.PostgresqlStandalone in the context.
type InstanceKey struct{}

// ConfigKey identifies the v1alpha1.PostgresqlStandaloneOperatorConfig in the context.
type ConfigKey struct{}

// HelmReleaseKey identifies the HelmRelease in the context.
type HelmReleaseKey struct{}

// DeploymentNamespaceKey identifies the deployment Namespace in the context.
type DeploymentNamespaceKey struct{}

// ServiceKey identifies the PostgreSQL service in the context.
type ServiceKey struct{}

// BucketSecretKey identifies the S3 bucket for Backup in the context.
type BucketSecretKey struct{}

// ConnectionSecretKey identifies the connection secret in the context.
type ConnectionSecretKey struct{}

// CredentialSecretKey identifies the credential secret for PostgreSQL in the context.
type CredentialSecretKey struct{}

// SetClientInContext sets the given client in the context.
func SetClientInContext(ctx context.Context, c client.Client) {
	pipeline.StoreInContext(ctx, ClientKey{}, c)
}

// GetClientFromContext returns the client from the context.
func GetClientFromContext(ctx context.Context) client.Client {
	return getFromContextOrPanic(ctx, ClientKey{}).(client.Client)
}

// SetInstanceInContext sets the given instance in the context.
func SetInstanceInContext(ctx context.Context, obj *v1alpha1.PostgresqlStandalone) {
	pipeline.StoreInContext(ctx, InstanceKey{}, obj)
}

// GetInstanceFromContext returns the instance from the context.
func GetInstanceFromContext(ctx context.Context) *v1alpha1.PostgresqlStandalone {
	return getFromContextOrPanic(ctx, InstanceKey{}).(*v1alpha1.PostgresqlStandalone)
}

// GetConfigFromContext returns the config from the context.
func GetConfigFromContext(ctx context.Context) *v1alpha1.PostgresqlStandaloneOperatorConfig {
	return getFromContextOrPanic(ctx, ConfigKey{}).(*v1alpha1.PostgresqlStandaloneOperatorConfig)
}

func getFromContextOrPanic(ctx context.Context, key any) any {
	val, exists := pipeline.LoadFromContext(ctx, key)
	if !exists {
		keyName := reflect.TypeOf(key).Name()
		panic(fmt.Errorf("key %q does not exist in the given context", keyName))
	}
	return val
}
