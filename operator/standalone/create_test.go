//go:build integration
// +build integration

package standalone

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateStandalonePipelineSuite struct {
	operatortest.Suite
}

func TestCreateStandalonePipeline(t *testing.T) {
	suite.Run(t, new(CreateStandalonePipelineSuite))
}

func (ts *CreateStandalonePipelineSuite) BeforeTest(suiteName, testName string) {
	ts.Ctx = context.Background()
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
				cfg := ts.newPostgresqlStandaloneOperatorConfig("config", "single-entry")
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
				cfg1 := ts.newPostgresqlStandaloneOperatorConfig("first", "multiple-entries")
				cfg1.Labels = map[string]string{
					v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
				}
				cfg2 := ts.newPostgresqlStandaloneOperatorConfig("second", "multiple-entries")
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
				instance:          ts.newInstance(),
			}
			tc.prepare()
			err := p.FetchOperatorConfig(ts.Ctx)
			if tc.expectedError != "" {
				ts.Require().EqualError(err, tc.expectedError, "fetch operator config")
				ts.Assert().Nil(p.config)
				return
			}
			ts.Assert().NoError(err)
		})
	}
}

func (ts *CreateStandalonePipelineSuite) newPostgresqlStandaloneOperatorConfig(name string, namespace string) *v1alpha1.PostgresqlStandaloneOperatorConfig {
	return &v1alpha1.PostgresqlStandaloneOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.PostgresqlMajorVersionLabelKey: v1alpha1.PostgresqlVersion14.String(),
			},
		},
	}
}
func (ts *CreateStandalonePipelineSuite) newInstance() *v1alpha1.PostgresqlStandalone {
	return &v1alpha1.PostgresqlStandalone{
		Spec: v1alpha1.PostgresqlStandaloneSpec{
			Parameters: v1alpha1.PostgresqlStandaloneParameters{
				MajorVersion: v1alpha1.PostgresqlVersion14,
			},
		},
	}
}
