package standalone

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const finalizer = "finalizer"

// SetupController adds a controller that reconciles v1alpha1.PostgresqlStandalone managed resources.
func SetupController(mgr ctrl.Manager, o controller.Options) error {
	name := strings.ToLower(v1alpha1.PostgresStandaloneGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.PostgresqlStandalone{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(&PostgresStandaloneReconciler{
			log:    o.Log,
			client: mgr.GetClient(),
		})
}

// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones/status;postgresqlstandalones/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update
// +kubebuilder:rbac:groups=helm.crossplane.io,resources=releases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=helm.crossplane.io,resources=providerconfigs,verbs=get;list;watch

// PostgresStandaloneReconciler reconciles v1alpha1.PostgresqlStandalone.
type PostgresStandaloneReconciler struct {
	client client.Client
	log    logr.Logger
}

// Reconcile implements reconcile.Reconciler.
func (r *PostgresStandaloneReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	obj := &v1alpha1.PostgresqlStandalone{}
	r.log.V(1).Info("Reconciling", "res", request.Name)
	err := r.client.Get(ctx, request.NamespacedName, obj)
	if err != nil && apierrors.IsNotFound(err) {
		// doesn't exist anymore, nothing to do
		return reconcile.Result{}, nil
	}
	if err != nil {
		// some other error
		return reconcile.Result{}, err
	}
	if !controllerutil.ContainsFinalizer(obj, finalizer) {
		return r.Create(ctx, obj)
	}
	if !obj.DeletionTimestamp.IsZero() {
		return r.Delete(ctx, obj)
	}
	return r.Update(ctx, obj)
}

// Create creates the given instance.
func (r *PostgresStandaloneReconciler) Create(ctx context.Context, instance *v1alpha1.PostgresqlStandalone) (reconcile.Result, error) {
	controllerutil.AddFinalizer(instance, finalizer)

	r.log.Info("Creating", "res", instance.Name)
	// also add some status condition here
	instance.Status.SetObservedGeneration(instance.ObjectMeta)
	err := r.client.Status().Update(ctx, instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.client.Update(ctx, instance)
	return reconcile.Result{}, err
}

// Delete prepares the given instance for deletion.
func (r *PostgresStandaloneReconciler) Delete(ctx context.Context, instance *v1alpha1.PostgresqlStandalone) (reconcile.Result, error) {
	// we don't need to delete it by ourselves, since the deletion timestamp is already set.
	// Just remove all finalizers
	controllerutil.RemoveFinalizer(instance, finalizer)
	r.log.Info("Deleting", "res", instance.Name)
	err := r.client.Update(ctx, instance)
	return reconcile.Result{}, err
}

// Update saves the given spec in Kubernetes.
func (r *PostgresStandaloneReconciler) Update(ctx context.Context, instance *v1alpha1.PostgresqlStandalone) (reconcile.Result, error) {
	r.log.Info("Updating", "res", instance.Name)
	// ensure status conditions are up-to-date.
	instance.Status.SetObservedGeneration(instance.ObjectMeta)
	err := r.client.Status().Update(ctx, instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.client.Update(ctx, instance)
	return reconcile.Result{}, err
}
