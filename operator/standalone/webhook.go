package standalone

import (
	"context"
	"fmt"

	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetupWebhook adds a webhook for v1alpha1.PostgresqlStandalone managed resources.
func SetupWebhook(mgr ctrl.Manager) error {
	/*
		Totally undocumented and hard-to-find feature is that the builder automatically registers the URL path for the webhook.
		What's more, not even the tests in upstream controller-runtime reveal what this path is _actually_ going to look like.
		So here's how the path is built (dots replaced with dash, lower-cased, single-form):
		 /validate-<group>-<version>-<kind>
		 /mutate-<group>-<version>-<kind>
		Example:
		 /validate-postgresql-appcat-vshn-io-v1alpha1-postgresqlstandalone
		This path has to be given in the `//+kubebuilder:webhook:...` magic comment, see example:
		 +kubebuilder:webhook:verbs=create;update;delete,path=/validate-postgresql-appcat-vshn-io-v1alpha1-postgresqlstandalone,mutating=false,failurePolicy=fail,groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones,versions=v1alpha1,name=postgresqlstandalones.postgresql.appcat.vshn.io,sideEffects=None,admissionReviewVersions=v1
		Pay special attention to the plural forms and correct versions!
	*/
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.PostgresqlStandalone{}).
		WithValidator(&PostgresqlStandaloneValidator{
			kube: mgr.GetClient(),
		}).
		Complete()
}

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
