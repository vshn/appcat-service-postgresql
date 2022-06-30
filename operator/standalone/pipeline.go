package standalone

import (
	"context"
	"fmt"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	crossplanev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// fetchHelmRelease fetches the Helm release for the given instance.
func fetchHelmRelease(ctx context.Context) error {
	helmRelease := &helmv1beta1.Release{}
	err := getClientFromContext(ctx).Get(ctx, client.ObjectKey{Name: getInstanceFromContext(ctx).Status.HelmChart.DeploymentNamespace}, helmRelease)
	setHelmReleaseInContext(ctx, helmRelease)
	return err
}

// fetchOperatorConfigF fetches a matching v1alpha1.PostgresqlStandaloneOperatorConfig from the OperatorNamespace.
// The Major version specified in v1alpha1.PostgresqlStandalone is used to filter the correct config by the v1alpha1.PostgresqlMajorVersionLabelKey label.
// If there is none or multiple found, it returns an error.
func fetchOperatorConfigF(operatorNamespace string) func(ctx2 context.Context) error {
	return func(ctx context.Context) error {
		list := &v1alpha1.PostgresqlStandaloneOperatorConfigList{}
		labelMatch := client.MatchingLabels{
			v1alpha1.PostgresqlMajorVersionLabelKey: getInstanceFromContext(ctx).Spec.Parameters.MajorVersion.String(),
		}
		ns := client.InNamespace(operatorNamespace)
		err := getClientFromContext(ctx).List(ctx, list, labelMatch, ns)
		if err != nil {
			return err
		}
		if len(list.Items) == 0 {
			return fmt.Errorf("no %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labelMatch, ns)
		}
		if len(list.Items) > 1 {
			return fmt.Errorf("multiple versions of %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labelMatch, ns)
		}
		setConfigInContext(ctx, &list.Items[0])
		return nil
	}
}

// applyValuesFromInstance merges the user-defined and -exposed Helm values into the current Helm values map.
func applyValuesFromInstance(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
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
	helmValues := getHelmValuesFromContext(ctx)
	helmvalues.Merge(resources, &helmValues)
	setHelmValuesInContext(ctx, helmValues)
	return nil
}

// ensureHelmRelease updates the Helm release object
func ensureHelmRelease(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	config := getConfigFromContext(ctx)
	chart := getHelmChartFromContext(ctx)
	helmValues, err := helmvalues.Marshal(getHelmValuesFromContext(ctx))
	if err != nil {
		return err
	}
	helmRelease := &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Status.HelmChart.DeploymentNamespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, getClientFromContext(ctx), helmRelease, func() error {
		helmRelease.Labels = labels.Merge(helmRelease.Labels, getCommonLabels(getInstanceFromContext(ctx).Name))
		helmRelease.Spec = helmv1beta1.ReleaseSpec{
			ForProvider: helmv1beta1.ReleaseParameters{
				Chart:               helmv1beta1.ChartSpec{Repository: chart.Repository, Name: chart.Name, Version: chart.Version},
				Namespace:           instance.Status.HelmChart.DeploymentNamespace,
				SkipCreateNamespace: true,
				SkipCRDs:            true,
				Wait:                true,
				ValuesSpec:          helmv1beta1.ValuesSpec{Values: helmValues},
			},
			ResourceSpec: crossplanev1.ResourceSpec{
				ProviderConfigReference: &crossplanev1.Reference{Name: config.Spec.HelmProviderConfigReference},
			},
		}
		return nil
	})
	setHelmReleaseInContext(ctx, helmRelease)
	return err
}

// overrideTemplateValues searches for a specific HelmRelease spec that matches the Chart version from the template spec.
// If it does, the template values are replaced or merged.
//
// This step assumes that the config has been fetched first via fetchOperatorConfigF.
func overrideTemplateValues(ctx context.Context) error {
	config := getConfigFromContext(ctx)
	helmValues := getHelmValuesFromContext(ctx)
	helmChart := getHelmChartFromContext(ctx)
	for _, release := range config.Spec.HelmReleases {
		// TODO: maybe a better semver comparison later on?
		if release.Chart.Version == config.Spec.HelmReleaseTemplate.Chart.Version {
			overrides := helmvalues.V{}
			err := helmvalues.Unmarshal(release.Values, &overrides)
			if err != nil {
				return err
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
	setHelmValuesInContext(ctx, helmValues)
	return nil
}

// useTemplateValues copies the Helm values and Chart metadata from the v1alpha1.PostgresqlStandaloneOperatorConfig spec as the starting parameters.
//
// This step assumes that the config has been fetched first via fetchOperatorConfigF.
func useTemplateValues(ctx context.Context) error {
	values := helmvalues.V{}
	config := getConfigFromContext(ctx)
	err := helmvalues.Unmarshal(config.Spec.HelmReleaseTemplate.Values, &values)
	setHelmValuesInContext(ctx, values)
	setHelmChartInContext(ctx, &config.Spec.HelmReleaseTemplate.Chart)
	return err
}

// ensurePVC ensures that the PVC is created.
//
// This step assumes that the deployment namespace name has been updated in the status of the instance.
func ensurePVC(ctx context.Context) error {
	instance := getInstanceFromContext(ctx)
	config := getConfigFromContext(ctx)
	persistentVolumeClaim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getPVCName(),
			Namespace: instance.Status.HelmChart.DeploymentNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, getClientFromContext(ctx), persistentVolumeClaim, func() error {
		persistentVolumeClaim.Labels = labels.Merge(persistentVolumeClaim.Labels, getCommonLabels(instance.Name))
		if len(persistentVolumeClaim.Spec.AccessModes) == 0 {
			persistentVolumeClaim.Spec.AccessModes = config.Spec.Persistence.AccessModes
		}
		if persistentVolumeClaim.Spec.StorageClassName == nil {
			persistentVolumeClaim.Spec.StorageClassName = config.Spec.Persistence.StorageClassName
		}
		persistentVolumeClaim.Spec.Resources = corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceStorage: *instance.Spec.Parameters.Resources.StorageCapacity},
		}
		return nil
	})
	return err
}
