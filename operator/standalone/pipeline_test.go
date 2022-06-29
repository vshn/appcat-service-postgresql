package standalone

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/helmvalues"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestCreateStandalonePipeline_UseTemplateValues(t *testing.T) {
	ctx := pipeline.MutableContext(context.Background())
	setConfigInContext(ctx, &v1alpha1.PostgresqlStandaloneOperatorConfig{Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
		HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
			Values: runtime.RawExtension{Raw: []byte(`{"key":"value"}`)},
			Chart:  v1alpha1.ChartMeta{Repository: "https://host/path", Name: "postgresql", Version: "1.0"},
		},
	}})
	err := useTemplateValues(ctx)
	assert.NoError(t, err)
	expectedValues := helmvalues.V{
		"key": "value",
	}
	expectedChart := &v1alpha1.ChartMeta{
		Repository: "https://host/path",
		Name:       "postgresql",
		Version:    "1.0",
	}
	assert.Equal(t, expectedValues, getHelmValuesFromContext(ctx))
	assert.Equal(t, expectedChart, getHelmChartFromContext(ctx))
}

func TestOverrideTemplateValues(t *testing.T) {
	tests := map[string]struct {
		givenSpec      v1alpha1.PostgresqlStandaloneOperatorConfigSpec
		expectedValues helmvalues.V
		expectedChart  v1alpha1.ChartMeta
		expectedError  string
	}{
		"GivenSpecificReleaseExists_WhenMergeEnabled_ThenMergeWithExistingValues": {
			givenSpec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
				HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
					Values: helmvalues.MustMarshal(helmvalues.V{"key": "value", "existing": "untouched"}),
					Chart:  v1alpha1.ChartMeta{Version: "version"},
				},
				HelmReleases: []v1alpha1.HelmReleaseConfig{
					{
						MergeValuesFromTemplate: true,
						Chart:                   v1alpha1.ChartMeta{Version: "version"},
						Values:                  helmvalues.MustMarshal(helmvalues.V{"key": map[string]interface{}{"nested": "value"}, "merged": "newValue"}),
					},
				},
			},
			expectedValues: helmvalues.V{
				"key": map[string]interface{}{
					"nested": "value",
				},
				"merged":   "newValue",
				"existing": "untouched",
			},
			expectedChart: v1alpha1.ChartMeta{Repository: "url", Name: "postgres", Version: "version"},
		},
		"GivenSpecificReleaseExists_WhenMergeDisabled_ThenOverwriteExistingValues": {
			givenSpec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
				HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
					Values: helmvalues.MustMarshal(helmvalues.V{"key": "value", "existing": "untouched"}),
					Chart:  v1alpha1.ChartMeta{Version: "version"},
				},
				HelmReleases: []v1alpha1.HelmReleaseConfig{
					{
						Chart:  v1alpha1.ChartMeta{Version: "version", Name: "alternative", Repository: "fork"},
						Values: helmvalues.MustMarshal(helmvalues.V{"key": helmvalues.V{"nested": "value"}, "merged": "newValue"}),
					},
				},
			},
			expectedValues: helmvalues.V{
				"key": map[string]interface{}{
					"nested": "value",
				},
				"merged": "newValue",
			},
			expectedChart: v1alpha1.ChartMeta{Repository: "fork", Name: "alternative", Version: "version"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			vals := helmvalues.V{}
			ctx := pipeline.MutableContext(context.Background())
			helmvalues.MustUnmarshal(tc.givenSpec.HelmReleaseTemplate.Values, &vals)
			setConfigInContext(ctx, &v1alpha1.PostgresqlStandaloneOperatorConfig{Spec: tc.givenSpec})
			setHelmValuesInContext(ctx, vals)
			setHelmChartInContext(ctx, &v1alpha1.ChartMeta{Repository: "url", Name: "postgres", Version: "version"})
			err := overrideTemplateValues(ctx)
			if tc.expectedError != "" {
				require.EqualError(t, err, tc.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedValues, getHelmValuesFromContext(ctx))
			assert.Equal(t, &tc.expectedChart, getHelmChartFromContext(ctx))
		})
	}
}

func TestApplyValuesFromInstance(t *testing.T) {
	ctx := pipeline.MutableContext(context.Background())
	setConfigInContext(ctx, newPostgresqlStandaloneOperatorConfig("cfg", "postgresql-system"))
	setInstanceInContext(ctx, newInstance("instance", "my-app"))
	setHelmValuesInContext(ctx, helmvalues.V{})
	err := applyValuesFromInstance(ctx)
	require.NoError(t, err)
	assert.Equal(t, helmvalues.V{
		"auth": helmvalues.V{
			"existingSecret":     "postgresql-credentials",
			"database":           "instance",
			"enablePostgresUser": true,
			"username":           "instance",
		},
		"primary": helmvalues.V{
			"persistence": helmvalues.V{
				"existingClaim": "postgresql-data",
			},
			"resources": helmvalues.V{
				"limits": helmvalues.V{
					"memory": "2Gi",
				},
			},
			"podAnnotations": helmvalues.V{
				"k8up.io/backupcommand":                      `sh -c 'PGUSER="postgres" PGPASSWORD="$POSTGRES_POSTGRES_PASSWORD" pg_dumpall --clean'`,
				"k8up.io/file-extension":                     ".sql",
				"postgresql.appcat.vshn.io/storage-capacity": "1Gi",
			},
		},
		"fullnameOverride": "postgresql",
		"networkPolicy": helmvalues.V{
			"enabled": true,
			"ingressRules": helmvalues.V{
				"primaryAccessOnlyFrom": helmvalues.V{
					"enabled": true,
					"namespaceSelector": helmvalues.V{
						"kubernetes.io/metadata.name": "my-app",
					},
				},
			},
		},
	}, getHelmValuesFromContext(ctx))
}

func newPostgresqlStandaloneOperatorConfig(name string, namespace string) *v1alpha1.PostgresqlStandaloneOperatorConfig {
	return &v1alpha1.PostgresqlStandaloneOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
			},
		},
		Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
			HelmProviderConfigReference: "helm-provider",
		},
	}
}
func parseResource(value string) *resource.Quantity {
	parsed := resource.MustParse(value)
	return &parsed
}

type PostgresqlStandaloneBuilder struct {
	*v1alpha1.PostgresqlStandalone
}

func newInstanceBuilder(name, namespace string) *PostgresqlStandaloneBuilder {
	return &PostgresqlStandaloneBuilder{newInstance(name, namespace)}
}

func newInstance(name string, namespace string) *v1alpha1.PostgresqlStandalone {
	return &v1alpha1.PostgresqlStandalone{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Generation: 1},
		Spec: v1alpha1.PostgresqlStandaloneSpec{
			Parameters: v1alpha1.PostgresqlStandaloneParameters{
				MajorVersion:    v1alpha1.PostgresqlVersion14,
				EnableSuperUser: true,
				Resources: v1alpha1.Resources{
					ComputeResources: v1alpha1.ComputeResources{MemoryLimit: parseResource("2Gi")},
					StorageResources: v1alpha1.StorageResources{StorageCapacity: parseResource("1Gi")},
				},
			},
		},
		Status: v1alpha1.PostgresqlStandaloneStatus{},
	}
}

func (b *PostgresqlStandaloneBuilder) setDeploymentNamespace(namespace string) *PostgresqlStandaloneBuilder {
	b.Status.PostgresqlStandaloneObservation = v1alpha1.PostgresqlStandaloneObservation{
		HelmChart: &v1alpha1.ChartMetaStatus{
			DeploymentNamespace: namespace,
		},
	}
	return b
}

func (b *PostgresqlStandaloneBuilder) setConnectionSecret(secret string) *PostgresqlStandaloneBuilder {
	b.Spec.ConnectableInstance = v1alpha1.ConnectableInstance{
		WriteConnectionSecretToRef: v1alpha1.ConnectionSecretRef{
			Name: secret,
		},
	}
	return b
}

func (b *PostgresqlStandaloneBuilder) setGenerationStatus(status v1alpha1.GenerationStatus) *PostgresqlStandaloneBuilder {
	b.Status.GenerationStatus = status
	return b
}

func (b *PostgresqlStandaloneBuilder) setConditions(conditions ...metav1.Condition) *PostgresqlStandaloneBuilder {
	b.Status.Conditions = conditions
	return b
}

func (b *PostgresqlStandaloneBuilder) enableSuperUser() *PostgresqlStandaloneBuilder {
	b.Spec.Parameters.EnableSuperUser = true
	return b
}

func (b *PostgresqlStandaloneBuilder) disableSuperUser() *PostgresqlStandaloneBuilder {
	b.Spec.Parameters.EnableSuperUser = false
	return b
}

func (b *PostgresqlStandaloneBuilder) setBackup(enabled bool) *PostgresqlStandaloneBuilder {
	b.Spec.Backup.Enabled = enabled
	return b
}

func (b *PostgresqlStandaloneBuilder) get() *v1alpha1.PostgresqlStandalone {
	return b.PostgresqlStandalone
}
