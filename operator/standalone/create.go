package standalone

import (
	"context"
	"fmt"
	"strings"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CreateStandalonePipeline struct {
	operatorNamespace string
	client            client.Client

	instance   *v1alpha1.PostgresqlStandalone
	config     *v1alpha1.PostgresqlStandaloneOperatorConfig
	helmValues HelmValues
	helmChart  *v1alpha1.ChartMeta
}

func (p *CreateStandalonePipeline) runPipeline(ctx context.Context) error {
	return pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("fetch operator config", p.FetchOperatorConfig),
			pipeline.NewPipeline().WithSteps(
				pipeline.NewStepFromFunc("read template values", p.UseTemplateValues),
				pipeline.NewStepFromFunc("override template values", p.OverrideTemplateValues),
				pipeline.NewStepFromFunc("apply values from instance", p.ApplyValuesFromInstance),
			).AsNestedStep("compile helm values"),
			pipeline.NewStepFromFunc("ensure deployment namespace", p.EnsureDeploymentNamespace),
			pipeline.NewStepFromFunc("ensure credentials secret", p.EnsureCredentialsSecret),
		).
		RunWithContext(ctx).Err()
}

func (p *CreateStandalonePipeline) FetchOperatorConfig(ctx context.Context) error {
	list := &v1alpha1.PostgresqlStandaloneOperatorConfigList{}
	labels := client.MatchingLabels{
		v1alpha1.PostgresqlMajorVersionLabelKey: p.instance.Spec.Parameters.MajorVersion.String(),
	}
	ns := client.InNamespace(p.operatorNamespace)
	err := p.client.List(ctx, list, labels, ns)
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("no %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labels, ns)
	}
	if len(list.Items) > 1 {
		return fmt.Errorf("multiple versions of %s found with label '%s' in namespace '%s'", v1alpha1.PostgresqlStandaloneOperatorConfigKind, labels, ns)
	}
	p.config = &list.Items[0]
	return nil
}

func (p *CreateStandalonePipeline) UseTemplateValues(_ context.Context) error {
	values := HelmValues{}
	err := values.Unmarshal(p.config.Spec.HelmReleaseTemplate.Values)
	p.helmValues = values
	p.helmChart = &p.config.Spec.HelmReleaseTemplate.Chart
	return err
}

func (p *CreateStandalonePipeline) OverrideTemplateValues(_ context.Context) error {
	for _, release := range p.config.Spec.HelmReleases {
		// TODO: maybe a better semver comparison later on?
		if release.Chart.Version == p.config.Spec.HelmReleaseTemplate.Chart.Version {
			overrides := HelmValues{}
			err := overrides.Unmarshal(release.Values)
			if err != nil {
				return err
			}
			if release.MergeValuesFromTemplate {
				p.helmValues.MergeWith(overrides)
			} else {
				p.helmValues = overrides
			}
			if release.Chart.Name != "" {
				p.helmChart.Name = release.Chart.Name
			}
			if release.Chart.Repository != "" {
				p.helmChart.Repository = release.Chart.Repository
			}
		}
	}
	return nil
}

func (p *CreateStandalonePipeline) ApplyValuesFromInstance(_ context.Context) error {
	resources := HelmValues{
		"auth": HelmValues{
			"enablePostgresUser": p.instance.Spec.Parameters.EnableSuperUser,
			"existingSecret":     getCredentialSecretName(p.instance),
			"database":           p.instance.Name,
		},
		"primary": HelmValues{
			"resources": HelmValues{
				"limits": HelmValues{
					"memory": p.instance.Spec.Parameters.Resources.MemoryLimit.String(),
				},
			},
			"persistence": HelmValues{
				"size": p.instance.Spec.Parameters.Resources.StorageCapacity.String(),
			},
		},
	}
	p.helmValues.MergeWith(resources)
	return nil
}

func (p *CreateStandalonePipeline) EnsureCredentialsSecret(ctx context.Context) error {
	// https://github.com/bitnami/charts/tree/master/bitnami/postgresql
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getCredentialSecretName(p.instance),
			Namespace: getNamespaceForInstance(p.instance),
			Labels:    getCommonLabels(p.instance.Name),
		},
		StringData: map[string]string{
			"postgres-password":    generatePassword(),
			"password":             generatePassword(),
			"replication-password": generatePassword(),
		},
	}
	// TODO: Add OwnerReference to Crossplane's HelmRelease
	// Note: We cannot set the reference to the instance, since cross-namespace references aren't allowed.
	return Upsert(ctx, p.client, secret)
}

func (p *CreateStandalonePipeline) EnsureDeploymentNamespace(ctx context.Context) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: getNamespaceForInstance(p.instance),
			// TODO: Add APPUiO cloud organization label that identifies ownership.
			Labels: getCommonLabels(p.instance.Name),
		},
	}
	return Upsert(ctx, p.client, ns)
}

func getCredentialSecretName(obj client.Object) string {
	return fmt.Sprintf("%s-credentials", obj.GetName())
}

func getCommonLabels(instanceName string) map[string]string {
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
	return map[string]string{
		"app.kubernetes.io/instance":   instanceName,
		"app.kubernetes.io/managed-by": v1alpha1.Group,
		"app.kubernetes.io/created-by": fmt.Sprintf("controller-%s", strings.ToLower(v1alpha1.PostgresStandaloneKind)),
	}
}

func getNamespaceForInstance(obj client.Object) string {
	// TODO: ensure that name doesn't exceed 63 characters
	return fmt.Sprintf("%s%s-%s", ServiceNamespacePrefix, obj.GetNamespace(), obj.GetName())
}

func generatePassword() string {
	return rand.String(40)
}
