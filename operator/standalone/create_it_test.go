//go:build integration

package standalone

import (
	"context"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	"k8s.io/apimachinery/pkg/runtime"
)

type CreateStandalonePipelineSuite struct {
	operatortest.Suite
}

func TestCreateStandalonePipeline(t *testing.T) {
	suite.Run(t, new(CreateStandalonePipelineSuite))
}

func (ts *CreateStandalonePipelineSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
}

func (ts *CreateStandalonePipelineSuite) Test_FetchOperatorConfig() {
	tests := map[string]struct {
		prepare            func()
		givenNamespace     string
		expectedConfigName string
		expectedError      string
	}{
		"GivenNoExistingConfig_WhenFetching_ThenExpectError": {
			prepare:        func() {},
			givenNamespace: "nonexisting",
			expectedError:  "no PostgresqlStandaloneOperatorConfig found with label 'map[postgresql.appcat.vshn.io/major-version:v14]' in namespace 'nonexisting'",
		},
		"GivenExistingConfig_WhenLabelsMatch_ThenExpectSingleEntry": {
			prepare: func() {
				ts.EnsureNS("single-entry")
				cfg := newPostgresqlStandaloneOperatorConfig("config", "single-entry")
				cfg.Labels = map[string]string{
					v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
				}
				ts.EnsureResources(cfg)
			},
			givenNamespace:     "single-entry",
			expectedConfigName: "config",
		},
		"GivenMultipleExistingConfigs_WhenLabelsMatch_ThenExpectError": {
			prepare: func() {
				ts.EnsureNS("multiple-entries")
				cfg1 := newPostgresqlStandaloneOperatorConfig("first", "multiple-entries")
				cfg1.Labels = map[string]string{
					v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
				}
				cfg2 := newPostgresqlStandaloneOperatorConfig("second", "multiple-entries")
				cfg2.Labels = cfg1.Labels
				ts.EnsureResources(cfg1, cfg2)
			},
			givenNamespace: "multiple-entries",
			expectedError:  "multiple versions of PostgresqlStandaloneOperatorConfig found with label 'map[postgresql.appcat.vshn.io/major-version:v14]' in namespace 'multiple-entries'",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			p := &CreateStandalonePipeline{
				operatorNamespace: tc.givenNamespace,
				client:            ts.Client,
				instance:          newInstance(),
			}
			tc.prepare()
			err := p.FetchOperatorConfig(ts.Context)
			if tc.expectedError != "" {
				ts.Require().EqualError(err, tc.expectedError)
				ts.Assert().Nil(p.config)
				return
			}
			ts.Assert().NoError(err)
		})
	}
}

func (ts *CreateStandalonePipelineSuite) Test_UseTemplateValues() {
	p := &CreateStandalonePipeline{
		config: &v1alpha1.PostgresqlStandaloneOperatorConfig{Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
			HelmReleaseTemplate: &v1alpha1.HelmReleaseConfig{
				Values: runtime.RawExtension{Raw: []byte(`{"key":"value"}`)},
			},
		}},
	}
	err := p.UseTemplateValues(ts.Context)
	ts.Require().NoError(err)
	expected := HelmValues{
		"key": "value",
	}
	ts.Assert().Equal(expected, p.helmValues)
}
