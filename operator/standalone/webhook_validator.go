package standalone

import (
	"context"
	"fmt"

	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PostgresqlStandaloneValidator validates admission requests.
type PostgresqlStandaloneValidator struct {
	kube client.Client
}

// ValidateCreate implements admission.CustomValidator.
func (v *PostgresqlStandaloneValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	res := obj.(*v1alpha1.PostgresqlStandalone)
	log := ctrl.LoggerFrom(ctx)
	log.V(1).Info("Validate create", "name", res.Name)
	//TODO implement me
	return nil
}

// ValidateUpdate implements admission.CustomValidator.
// This validator:
//  - prevents selecting another major version (major version upgrade is currently unsupported)
//  - prevents storage capacity to be decreased
func (v *PostgresqlStandaloneValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) error {
	newInstance := newObj.(*v1alpha1.PostgresqlStandalone)
	oldInstance := oldObj.(*v1alpha1.PostgresqlStandalone)
	if newInstance.Spec.Parameters.MajorVersion != oldInstance.Spec.Parameters.MajorVersion {
		return fmt.Errorf("major version cannot be changed once specified at creation time")
	}
	if newInstance.Spec.Parameters.Resources.StorageCapacity.Cmp(*oldInstance.Spec.Parameters.Resources.StorageCapacity) == -1 {
		return fmt.Errorf("storage capacity cannot be decreased")
	}
	return nil
}

// ValidateDelete implements admission.CustomValidator.
func (v *PostgresqlStandaloneValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	res := obj.(*v1alpha1.PostgresqlStandalone)
	log := ctrl.LoggerFrom(ctx)
	log.V(1).Info("Validate delete", "name", res.Name)
	//TODO implement me
	return nil
}
