package standalone

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/operator/steps"
)

// DeleteStandalonePipeline is a pipeline that deletes an instance from the target deployment namespace.
type DeleteStandalonePipeline struct{}

// NewDeleteStandalonePipeline creates a new delete pipeline with the required dependencies.
func NewDeleteStandalonePipeline() *DeleteStandalonePipeline {
	return &DeleteStandalonePipeline{}
}

// RunPipeline executes the pipeline with configured business logic steps.
// The pipeline requires multiple reconciliations due to asynchronous deletion of resources in background
// The Helm Release step requires a complete removal of its resources before moving to the next step
func (d *DeleteStandalonePipeline) RunPipeline(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("delete connection secret", steps.DeleteConnectionSecretFn()),
			pipeline.NewStepFromFunc("delete helm release", steps.DeleteHelmReleaseFn()),
			pipeline.NewStepFromFunc("delete pvc", steps.DeletePvcFn()),
			pipeline.NewStepFromFunc("delete namespace", steps.DeleteNamespaceFn()),
			pipeline.NewStepFromFunc("remove finalizer", steps.RemoveFinalizerFn(finalizer)),
		).
		RunWithContext(ctx).Err()
}
