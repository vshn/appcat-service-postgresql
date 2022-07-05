package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PostgresqlStandaloneObservation are the observable fields of a PostgresqlStandalone.
type PostgresqlStandaloneObservation struct {
	// DeploymentStrategy is the observed deployed strategy.
	DeploymentStrategy DeploymentStrategy `json:"deploymentStrategy,omitempty"`
	// HelmChart is the observed deployed Helm chart version.
	HelmChart *ChartMetaStatus `json:"helmChart,omitempty"`
}

type GenerationStatus struct {
	// ObservedGeneration is the meta.generation number this resource was last reconciled with.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// SetObservedGeneration sets the ObservedGeneration from the given ObjectMeta.
func (in *GenerationStatus) SetObservedGeneration(obj client.Object) {
	in.ObservedGeneration = obj.GetGeneration()
}

// GetDeploymentNamespace returns the name of the namespace where the instance is deployed.
func (in PostgresqlStandaloneObservation) GetDeploymentNamespace() string {
	if in.HelmChart == nil {
		return ""
	}
	return in.HelmChart.DeploymentNamespace
}
