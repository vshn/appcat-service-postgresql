package steps

import (
	"context"
	"fmt"
	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// EnsureHelmReleaseFn creates or updates the Helm release object.
// For first time installations, the Helm values are compiled based on the v1alpha1.PostgresqlStandaloneOperatorConfig HelmReleaseTemplate.
// For updates, the existing Helm values are merged with values that are specific to the instance.
// A release is considered "new" if the v1alpha1.PostgresqlStandalone's Status.HelmChart is nil.
func EnsureHelmReleaseFn(labelSet labels.Set) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)
		config := GetConfigFromContext(ctx)
		deploymentNamespace := getFromContextOrPanic(ctx, DeploymentNamespaceKey{}).(*corev1.Namespace)

		helmRelease := &helmv1beta1.Release{ObjectMeta: metav1.ObjectMeta{Name: deploymentNamespace.Name}}
		_, err := controllerutil.CreateOrUpdate(ctx, kube, helmRelease, func() error {
			chart := helmRelease.Spec.ForProvider.Chart
			values := helmvalues.V{}
			if helmRelease.ResourceVersion == "" {
				// new instance.
				compiledValues, chartSpec, err := compileHelmValues(config, instance)
				if err != nil {
					return err
				}
				values = compiledValues
				chart.Repository = chartSpec.Repository
				chart.Name = chartSpec.Name
				chart.Version = chartSpec.Version
			} else {
				// existing release.
				// Due to the delayable maintenance feature coming up, we can't compile the Helm values from template,
				//  as that would potentially change values like image tags and thus do an unscheduled update of instance even in cases where the user just wanted more memory.
				// But that means that we assume that no human is playing around with Helm values in the Release spec, since we rely on the existing Helm values to be valid.
				// So we use the existing values and merge with the values directly set by the instance.
				existingValues := helmvalues.V{}
				ext := helmRelease.Spec.ForProvider.Values
				instance.Status.HelmChart.SetHashSumOfExistingValues(helmvalues.MustHashSum(helmRelease.Spec.ForProvider.Values))
				err := helmvalues.Unmarshal(ext, &existingValues)
				if err != nil {
					return err
				}
				values = applyValuesFromInstance(instance, existingValues)
			}
			rawExt, err := helmvalues.Marshal(values)
			if err != nil {
				return err
			}

			helmRelease.Labels = labels.Merge(helmRelease.Labels, labelSet)
			helmRelease.Spec = helmv1beta1.ReleaseSpec{
				ForProvider: helmv1beta1.ReleaseParameters{
					Chart:               chart,
					Namespace:           deploymentNamespace.Name,
					SkipCreateNamespace: true,
					SkipCRDs:            true,
					Wait:                true,
					ValuesSpec:          helmv1beta1.ValuesSpec{Values: rawExt},
				},
				ResourceSpec: crossplanev1.ResourceSpec{
					ProviderConfigReference: &crossplanev1.Reference{Name: config.Spec.HelmProviderConfigReference},
				},
			}
			return nil
		})
		pipeline.StoreInContext(ctx, HelmReleaseKey{}, helmRelease)
		return err
	}
}

func compileHelmValues(config *v1alpha1.PostgresqlStandaloneOperatorConfig, instance *v1alpha1.PostgresqlStandalone) (helmvalues.V, *v1alpha1.ChartMeta, error) {
	helmVals := helmvalues.V{}

	err := helmvalues.Unmarshal(config.Spec.HelmReleaseTemplate.Values, &helmVals)
	if err != nil {
		return nil, nil, err
	}
	helmVals, helmChart, err := overrideTemplateValues(config, helmVals)
	if err != nil {
		return nil, nil, err
	}

	helmVals = applyValuesFromInstance(instance, helmVals)
	return helmVals, helmChart, nil
}

func getCredentialSecretName() string {
	return fmt.Sprintf("%s-credentials", getDeploymentName())
}
func getDeploymentName() string {
	return "postgresql"
}
func getPVCName() string {
	return "postgresql-data"
}

// overrideTemplateValues searches for a specific HelmRelease spec that matches the Chart version from the template spec.
// If it does, the template values are replaced or merged.
func overrideTemplateValues(config *v1alpha1.PostgresqlStandaloneOperatorConfig, helmValues helmvalues.V) (helmvalues.V, *v1alpha1.ChartMeta, error) {
	helmChart := &config.Spec.HelmReleaseTemplate.Chart

	for _, release := range config.Spec.HelmReleases {
		// TODO: maybe a better semver comparison later on?
		if release.Chart.Version == config.Spec.HelmReleaseTemplate.Chart.Version {
			overrides := helmvalues.V{}
			err := helmvalues.Unmarshal(release.Values, &overrides)
			if err != nil {
				return helmValues, helmChart, err
			}
			if release.MergeValuesFromTemplate {
				helmvalues.Merge(overrides, &helmValues)
			} else {
				helmValues = overrides
			}
			if release.Chart.Name != "" {
				helmChart.Name = release.Chart.Name
			}
			if release.Chart.Repository != "" {
				helmChart.Repository = release.Chart.Repository
			}
		}
	}
	return helmValues, helmChart, nil
}

// applyValuesFromInstance merges the user-defined and -exposed Helm values into the current Helm values map.
func applyValuesFromInstance(instance *v1alpha1.PostgresqlStandalone, values helmvalues.V) helmvalues.V {
	resources := helmvalues.V{
		"auth": helmvalues.V{
			"enablePostgresUser": true, // See https://github.com/vshn/appcat-service-postgresql/issues/83 why we always create a superuser
			"existingSecret":     getCredentialSecretName(),
			"database":           instance.Name,
			"username":           instance.Name,
		},
		"primary": helmvalues.V{
			"resources": helmvalues.V{
				"limits": helmvalues.V{
					"memory": instance.Spec.Parameters.Resources.MemoryLimit.String(),
				},
			},
			"persistence": helmvalues.V{
				"existingClaim": getPVCName(),
			},
			"podAnnotations": helmvalues.V{ // these annotations can stay, even if backups are disabled.
				"k8up.io/backupcommand":                      `sh -c 'PGUSER="postgres" PGPASSWORD="$POSTGRES_POSTGRES_PASSWORD" pg_dumpall --clean'`,
				"k8up.io/file-extension":                     ".sql",
				"postgresql.appcat.vshn.io/storage-capacity": instance.Spec.Parameters.Resources.StorageCapacity.String(),
			},
		},
		"fullnameOverride": getDeploymentName(),
		"networkPolicy": helmvalues.V{
			"enabled": true,
			"ingressRules": helmvalues.V{
				"primaryAccessOnlyFrom": helmvalues.V{
					"enabled": true,
					"namespaceSelector": helmvalues.V{
						"kubernetes.io/metadata.name": instance.Namespace,
					},
				},
			},
		},
	}
	helmvalues.Merge(resources, &values)
	return values
}

// IsHelmReleaseReadyP returns a predicate that returns true if the HelmRelease has the ready condition.
func IsHelmReleaseReadyP() func(ctx context.Context) bool {
	return func(ctx context.Context) bool {
		instance := GetInstanceFromContext(ctx)
		helmRelease := getFromContextOrPanic(ctx, HelmReleaseKey{}).(*helmv1beta1.Release)

		if helmRelease.Status.Synced {
			if readyCondition := FindCrossplaneCondition(helmRelease.Status.Conditions, crossplanev1.TypeReady); readyCondition != nil && readyCondition.Status == corev1.ConditionTrue {
				modified := instance.Status.HelmChart.ModifiedTime
				lastReady := &readyCondition.LastTransitionTime
				// If ready condition was updated after our last saved modification time, then it must have been reconciled.
				return modified.Before(lastReady) || modified.Equal(lastReady)
			}
		}
		return false
	}
}

// EnrichStatusWithHelmChartMetaFn returns a function that updates the instance's status with metadata.
func EnrichStatusWithHelmChartMetaFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		instance := GetInstanceFromContext(ctx)
		kube := GetClientFromContext(ctx)
		helmRelease := getFromContextOrPanic(ctx, HelmReleaseKey{}).(*helmv1beta1.Release)

		helmChart := helmRelease.Spec.ForProvider.Chart
		if instance.Status.HelmChart == nil {
			instance.Status.HelmChart = &v1alpha1.ChartMetaStatus{}
		}
		instance.Status.HelmChart.ChartMeta = v1alpha1.ChartMeta{
			Repository: helmChart.Repository,
			Version:    helmChart.Version,
			Name:       helmChart.Name,
		}
		instance.Status.HelmChart.DeploymentNamespace = helmRelease.Spec.ForProvider.Namespace
		instance.Status.DeploymentStrategy = v1alpha1.StrategyHelmChart
		instance.Status.SetObservedGeneration(instance)
		valuesHash := helmvalues.MustHashSum(helmRelease.Spec.ForProvider.Values)
		if instance.Status.HelmChart.GetHashSumOfExistingValues() != valuesHash {
			instance.Status.HelmChart.ModifiedTime = metav1.Now()
		}
		err := kube.Status().Update(ctx, instance)
		return err
	}
}

// DeleteHelmReleaseFn removes the Helm Release from the cluster.
// Ignores "not found" error and returns nil if deployment namespace is unknown.
func DeleteHelmReleaseFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)

		if instance.Status.GetDeploymentNamespace() == "" {
			// Release might not ever have existed, skip
			return nil
		}
		helmRelease := &helmv1beta1.Release{
			ObjectMeta: metav1.ObjectMeta{
				Name: instance.Status.HelmChart.DeploymentNamespace,
			},
		}
		err := kube.Delete(ctx, helmRelease)
		return client.IgnoreNotFound(err)
	}
}

// FindCrossplaneCondition finds the conditionType in conditions.
func FindCrossplaneCondition(conditions []crossplanev1.Condition, conditionType crossplanev1.ConditionType) *crossplanev1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
