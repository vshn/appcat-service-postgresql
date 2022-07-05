package standalone

import (
	"context"
	"fmt"
	"github.com/vshn/appcat-service-postgresql/operator/steps"
	"k8s.io/apimachinery/pkg/labels"
	"math/rand"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"strings"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/lucasepe/codename"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
)

var namegeneratorRNG *rand.Rand

func init() {
	rng, err := codename.DefaultRNG()
	if err != nil {
		panic(err)
	}
	namegeneratorRNG = rng
}

// CreateStandalonePipeline is a pipeline that creates a new instance in the target deployment namespace.
// Currently, it's optimized for first-time creation scenarios and may fail when reconciling existing instances.
type CreateStandalonePipeline struct {
	operatorNamespace string
}

// NewStandalonePipeline creates a new pipeline with the required dependencies.
func NewStandalonePipeline(operatorNamespace string) *CreateStandalonePipeline {
	return &CreateStandalonePipeline{
		operatorNamespace: operatorNamespace,
	}
}

// Run executes the pipeline with configured business logic steps.
// This should only be executed once per pipeline as it stores intermediate results in the struct.
func (p *CreateStandalonePipeline) Run(ctx context.Context) error {
	instance := steps.GetInstanceFromContext(ctx)
	commonLabels := getCommonLabels(instance.Name)

	// TODO: Add APPUiO cloud organization label that identifies ownership.
	nsLabelSet := labels.Merge(commonLabels, labels.Set{"app.kubernetes.io/instance-namespace": instance.Namespace})

	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch operator config", steps.FetchOperatorConfigFn(p.operatorNamespace)),
			pipeline.NewStepFromFunc("fetch instance namespace", steps.FetchNamespaceFn(instance.Namespace, steps.InstanceNamespaceKey{})),

			pipeline.NewStepFromFunc("add finalizer", steps.AddFinalizerFn(instance, finalizer)),
			pipeline.NewStepFromFunc("mark instance as progressing", steps.MarkInstanceAsProgressingFn()),

			pipeline.NewPipeline().WithNestedSteps("deploy resources",
				pipeline.NewStepFromFunc("ensure deployment namespace", steps.EnsureNamespace(getDeploymentNamespaceOrGenerate(instance), nsLabelSet)),
				pipeline.NewStepFromFunc("ensure PVC", steps.EnsurePvcFn(commonLabels)),
				pipeline.NewStepFromFunc("ensure credentials secret", steps.EnsureCredentialsSecretFn(commonLabels)),
				pipeline.NewStepFromFunc("ensure helm release", steps.EnsureHelmReleaseFn(commonLabels)),
				pipeline.NewStepFromFunc("enrich status with chart meta", steps.EnrichStatusWithHelmChartMetaFn()),
				pipeline.IfOrElse(steps.IsBackupEnabledP(),
					pipeline.NewPipeline().WithNestedSteps("ensure backup",
						// TODO: add step to provision S3 bucket
						pipeline.NewStepFromFunc("fetch bucket secret", steps.FetchS3BucketSecretFn()),
						pipeline.NewStepFromFunc("ensure encryption secret", steps.EnsureResticRepositorySecretFn(commonLabels)),
						pipeline.NewStepFromFunc("ensure k8up schedule", steps.EnsureK8upScheduleFn(commonLabels)),
					),
					// else
					pipeline.NewStepFromFunc("delete k8up schedule", steps.DeleteK8upScheduleFn())),
			),

			pipeline.If(steps.IsHelmReleaseReadyP(),
				pipeline.NewPipeline().WithNestedSteps("finish provisioning",
					pipeline.NewPipeline().WithNestedSteps("create connection secret",
						pipeline.NewStepFromFunc("fetch service", steps.FetchServiceFn()),
						pipeline.NewStepFromFunc("ensure connection secret", steps.EnsureConnectionSecretFn(commonLabels)),
					),
					pipeline.NewStepFromFunc("mark instance ready", steps.MarkInstanceAsReadyFn()).WithResultHandler(p.logProvisioningFinished),
				),
			),
		).
		RunWithContext(ctx).Err()
}

func (p *CreateStandalonePipeline) logProvisioningFinished(ctx context.Context, result pipeline.Result) error {
	if result.IsSuccessful() {
		log := controllerruntime.LoggerFrom(ctx)
		log.Info("Provisioning finished")
	}
	return result.Err()
}

func getCommonLabels(instanceName string) labels.Set {
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
	return labels.Set{
		"app.kubernetes.io/instance":   instanceName,
		"app.kubernetes.io/managed-by": v1alpha1.Group,
		"app.kubernetes.io/created-by": fmt.Sprintf("controller-%s", strings.ToLower(v1alpha1.PostgresqlStandaloneKind)),
	}
}

func generateClusterScopedNameForInstance() string {
	name := ""
	for i := 0; i < 10; i++ {
		name = fmt.Sprintf("%s%s", ServiceNamespacePrefix, codename.Generate(namegeneratorRNG, 4))
		if len(name) <= 63 && !strings.HasSuffix(name, "-") {
			return name
		}
	}
	panic("no generated name under 63 chars without '-' suffix after 10 attempts, giving up")
}

func getDeploymentNamespaceOrGenerate(instance *v1alpha1.PostgresqlStandalone) string {
	if instance.Status.HelmChart != nil && instance.Status.HelmChart.DeploymentNamespace != "" {
		return instance.Status.HelmChart.DeploymentNamespace
	}
	return generateClusterScopedNameForInstance()
}
