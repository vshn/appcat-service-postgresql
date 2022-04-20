package standalone

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
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
	name := managed.ControllerName(v1alpha1.PostgresStandaloneGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.PostgresqlStandalone{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(&PostgresStandaloneReconciler{
			log:    o.Log,
			client: mgr.GetClient(),
		})
}

// +kubebuilder:rbac:groups=postgres.appcat.vshn.io,resources=postgresstandalones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgres.appcat.vshn.io,resources=postgresstandalones/status;postgresstandalones/finalizers,verbs=get;update;patch

type PostgresStandaloneReconciler struct {
	client client.Client
	log    logr.Logger
}

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

func (r *PostgresStandaloneReconciler) Delete(ctx context.Context, instance *v1alpha1.PostgresqlStandalone) (reconcile.Result, error) {
	// we don't need to delete it by ourselves, since the deletion timestamp is already set.
	// Just remove all finalizers
	controllerutil.RemoveFinalizer(instance, finalizer)
	r.log.Info("Deleting", "res", instance.Name)
	err := r.client.Update(ctx, instance)
	return reconcile.Result{}, err
}

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
