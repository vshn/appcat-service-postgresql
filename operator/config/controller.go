package config

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/vshn/appcat-service-postgresql/apis/provider/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Setup adds a controller that reconciles ProviderConfigs by accounting for their current usage.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ProviderConfig{}).
		Complete(&ProviderConfigReconciler{
			log:    o.Log,
			client: mgr.GetClient(),
		})
}

// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=providerconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=providerconfigs/status;providerconfigs/finalizers,verbs=get;update;patch

// ProviderConfigReconciler reconciles v1alpha1.ProviderConfig.
type ProviderConfigReconciler struct {
	client client.Client
	log    logr.Logger
}

// Reconcile implements reconcile.Reconciler.
func (r *ProviderConfigReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	obj := &v1alpha1.ProviderConfig{}
	r.log.V(1).Info("Reconciling", "res", obj.Name)
	return reconcile.Result{}, nil
}
