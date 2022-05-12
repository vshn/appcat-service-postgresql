package config

import (
	"context"

	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SetupController adds a controller that reconciles ProviderConfigs by accounting for their current usage.
func SetupController(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.PostgresqlStandaloneOperatorConfig{}).
		Complete(&PostgresqlStandaloneOperatorConfigReconciler{
			client: mgr.GetClient(),
		})
}

// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=postgresqlstandaloneoperatorconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=postgresqlstandaloneoperatorconfigs/status;postgresqlstandaloneoperatorconfigs/finalizers,verbs=get;update;patch

// PostgresqlStandaloneOperatorConfigReconciler reconciles v1alpha1.ProviderConfig.
type PostgresqlStandaloneOperatorConfigReconciler struct {
	client client.Client
}

// Reconcile implements reconcile.Reconciler.
func (r *PostgresqlStandaloneOperatorConfigReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	_ = &v1alpha1.PostgresqlStandaloneOperatorConfig{}
	log := ctrl.LoggerFrom(ctx)
	log.V(1).Info("Reconciling")
	return reconcile.Result{}, nil
}
