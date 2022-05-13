package standalone

import (
	"strings"

	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// SetupController adds a controller that reconciles v1alpha1.PostgresqlStandalone managed resources.
func SetupController(mgr ctrl.Manager) error {
	name := strings.ToLower(v1alpha1.PostgresqlStandaloneGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.PostgresqlStandalone{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(&PostgresStandaloneReconciler{
			client: mgr.GetClient(),
		})
}

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
