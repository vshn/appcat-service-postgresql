package standalone

import (
	"context"

	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PostgresqlStandaloneDefaulter is the webhook that sets default values for the v1alpha1.PostgresqlStandalone.
type PostgresqlStandaloneDefaulter struct{}

// Default sets the default values for the instance.
func (p *PostgresqlStandaloneDefaulter) Default(_ context.Context, obj runtime.Object) error {
	instance := obj.(*v1alpha1.PostgresqlStandalone)
	if instance.Spec.WriteConnectionSecretToRef.Name == "" {
		instance.Spec.WriteConnectionSecretToRef.Name = instance.Name
	}
	return nil
}
