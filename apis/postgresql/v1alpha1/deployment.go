package v1alpha1

// DeploymentStrategy refers to different backend implementation how the instance is being deployed in the background.
type DeploymentStrategy string

const (
	// StrategyHelmChart refers to a DeploymentStrategy that deploys the instance using a Helm chart.
	StrategyHelmChart DeploymentStrategy = "HelmChart"
)
