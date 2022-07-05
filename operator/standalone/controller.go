package standalone

import (
	"context"
	"github.com/vshn/appcat-service-postgresql/operator/steps"
	"strings"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var finalizer = strings.ToLower(strings.ReplaceAll(v1alpha1.PostgresqlStandaloneGroupKind, ".", "-"))

var (
	// OperatorNamespace is the namespace where the controller looks for v1alpha1.PostgresqlStandaloneOperatorConfig.
	OperatorNamespace = ""
	// ServiceNamespacePrefix is the namespace prefix which the controller uses to create the namespaces where the PostgreSQL instances are actually deployed in.
	ServiceNamespacePrefix = "sv-postgresql-s-"
)

// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=postgresqlstandalones/status;postgresqlstandalones/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=postgresqlstandaloneoperatorconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgresql.appcat.vshn.io,resources=postgresqlstandaloneoperatorconfigs/status;postgresqlstandaloneoperatorconfigs/finalizers,verbs=get;update;patch

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update
// +kubebuilder:rbac:groups=helm.crossplane.io,resources=releases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=helm.crossplane.io,resources=providerconfigs,verbs=get;list;watch
// +kubebuilder:rbac:groups=k8up.io,resources=schedules,verbs=get;list;watch;create;update;patch;delete

// PostgresStandaloneReconciler reconciles v1alpha1.PostgresqlStandalone.
type PostgresStandaloneReconciler struct {
	client client.Client
}

// Reconcile implements reconcile.Reconciler.
func (r *PostgresStandaloneReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	ctx = pipeline.MutableContext(ctx)
	steps.SetClientInContext(ctx, r.client)
	obj := &v1alpha1.PostgresqlStandalone{}
	steps.SetInstanceInContext(ctx, obj)
	log := ctrl.LoggerFrom(ctx)
	log.V(1).Info("Reconciling")
	err := r.client.Get(ctx, request.NamespacedName, obj)
	if err != nil && apierrors.IsNotFound(err) {
		// doesn't exist anymore, nothing to do
		return reconcile.Result{}, nil
	}
	if err != nil {
		// some other error
		return reconcile.Result{}, err
	}
	if !obj.DeletionTimestamp.IsZero() {
		return r.DeleteDeployment(ctx)
	}
	return r.ProvisionDeployment(ctx, obj)
}

// ProvisionDeployment reconciles the given instance
func (r *PostgresStandaloneReconciler) ProvisionDeployment(ctx context.Context, instance *v1alpha1.PostgresqlStandalone) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	p := NewStandalonePipeline(OperatorNamespace)
	log.Info("Provisioning instance")
	err := p.Run(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !meta.IsStatusConditionTrue(instance.Status.Conditions, conditions.TypeReady) {
		// The instance has provisioned all the resources, now we'll have to wait until everything is ready.
		log.Info("Waiting until instance becomes ready")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}
	return reconcile.Result{}, nil
}

// DeleteDeployment prepares the given instance for deletion.
func (r *PostgresStandaloneReconciler) DeleteDeployment(ctx context.Context) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	d := NewDeleteStandalonePipeline()
	log.Info("Deleting instance")
	err := d.RunPipeline(ctx)
	return reconcile.Result{RequeueAfter: 1 * time.Second}, err
}
