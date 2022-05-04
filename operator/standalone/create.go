package standalone

import (
	"context"
	"fmt"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CreateStandalonePipeline struct {
	operatorNamespace string
	client            client.Client

	instance   *v1alpha1.PostgresqlStandalone
	config     *v1alpha1.PostgresqlStandaloneOperatorConfig
	helmValues HelmValues
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
				err = p.helmValues.MergeWith(overrides)
				if err != nil {
					return err
				}
			} else {
				p.helmValues = overrides
			}
		}
	}
	return nil
}

func (p *CreateStandalonePipeline) ApplyValuesFromInstance(_ context.Context) error {
	resources := HelmValues{
		"auth": HelmValues{
			"enablePostgresUser": p.instance.Spec.Parameters.EnableSuperUser,
			"existingSecret":     fmt.Sprintf("%s-credentials", p.instance.Name),
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
	return p.helmValues.MergeWith(resources)
}
