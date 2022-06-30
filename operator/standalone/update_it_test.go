//go:build integration

package standalone

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/conditions"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type UpdateStandalonePipelineSuite struct {
	operatortest.Suite
}

func TestUpdateStandalonePipeline(t *testing.T) {
	suite.Run(t, new(UpdateStandalonePipelineSuite))
}

func (ts *UpdateStandalonePipelineSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	setClientInContext(ts.Context, ts.Client)
	ts.RegisterScheme(helmv1beta1.SchemeBuilder.AddToScheme)
	ts.RegisterScheme(k8upv1.SchemeBuilder.AddToScheme)
}

func (ts *UpdateStandalonePipelineSuite) Test_PatchConnectionSecret() {
	tests := map[string]struct {
		prepare                  func()
		givenInstance            *v1alpha1.PostgresqlStandalone
		givenNamespace           string
		expectedConnectionSecret *v1.Secret
	}{
		"GivenConnectionSecretWithoutSuperUserPassword_WhenEnableUserIsTrue_ThenExpectPostgresqlPasswordInSecret": {
			prepare: func() {
				ts.EnsureNS("connection-secret-namespace-one")
				ts.EnsureNS("deployment-namespace-one")
				credentialsSecret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "postgresql-credentials",
						Namespace: "deployment-namespace-one",
					},
					Data: map[string][]byte{
						"postgres-password": []byte("postgres-password"),
					},
				}
				connectionSecret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection-secret",
						Namespace: "connection-secret-namespace-one",
					},
					Data: map[string][]byte{
						"user-password": []byte("user-password"),
						"data":          []byte("data"),
					},
				}
				ts.EnsureResources(credentialsSecret, connectionSecret)
			},
			givenInstance: newInstanceBuilder("instance", "connection-secret-namespace-one").
				setConnectionSecret("connection-secret").
				setSuperUserEnabled(true).
				setDeploymentNamespace("deployment-namespace-one").
				getInstance(),
			givenNamespace: "connection-secret-namespace-one",
			expectedConnectionSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "connection-secret",
					Namespace: "connection-secret-namespace-one",
				},
				Data: map[string][]byte{
					"user-password":                []byte("user-password"),
					"data":                         []byte("data"),
					"POSTGRESQL_POSTGRES_PASSWORD": []byte("postgres-password"),
				},
			},
		},
		"GivenConnectionSecretWithSuperUserPassword_WhenEnableUserIsFalse_ThenRemovePostgresqlPasswordFromSecret": {
			prepare: func() {
				ts.EnsureNS("connection-secret-namespace-two")
				ts.EnsureNS("deployment-namespace-two")
				credentialsSecret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "postgresql-credentials",
						Namespace: "deployment-namespace-two",
					},
					Data: map[string][]byte{
						"postgres-password": []byte("postgres-password"),
					},
				}
				connectionSecret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection-secret",
						Namespace: "connection-secret-namespace-two",
					},
					Data: map[string][]byte{
						"user-password":                []byte("user-password"),
						"data":                         []byte("data"),
						"POSTGRESQL_POSTGRES_PASSWORD": []byte("postgres-password"),
					},
				}
				ts.EnsureResources(credentialsSecret, connectionSecret)
			},
			givenInstance: newInstanceBuilder("instance", "connection-secret-namespace-two").
				setConnectionSecret("connection-secret").
				setSuperUserEnabled(false).
				setDeploymentNamespace("deployment-namespace-two").
				getInstance(),
			givenNamespace: "connection-secret-namespace-two",
			expectedConnectionSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "connection-secret",
					Namespace: "connection-secret-namespace-two",
				},
				Data: map[string][]byte{
					"user-password": []byte("user-password"),
					"data":          []byte("data"),
				},
			},
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			p := UpdateStandalonePipeline{}
			setClientInContext(ts.Context, ts.Client)
			setInstanceInContext(ts.Context, tc.givenInstance)
			tc.prepare()

			// Act
			err := p.patchConnectionSecret(ts.Context)
			ts.Require().NoError(err)

			// Assert
			actualSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "connection-secret",
					Namespace: tc.givenNamespace,
				},
			}
			ts.FetchResource(client.ObjectKeyFromObject(actualSecret), actualSecret)
			ts.Assert().Equal(tc.expectedConnectionSecret.Data, actualSecret.Data)
		})
	}
}

func (ts *UpdateStandalonePipelineSuite) Test_MarkInstanceAsProgressing() {
	tests := map[string]struct {
		prepare          func(*v1alpha1.PostgresqlStandalone)
		givenInstance    *v1alpha1.PostgresqlStandalone
		expectedInstance *v1alpha1.PostgresqlStandalone
	}{
		"GivenInstanceIsUpdated_WhenInitialUpdateReconcileFinished_ThenExpectStatusProgressingTrue": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				ts.EnsureNS("instance-first-namespace")
				ts.EnsureResources(instance)
			},
			givenInstance: newInstance("my-first-instance", "instance-first-namespace"),
			expectedInstance: newInstanceBuilder("my-first-instance", "instance-first-namespace").
				setGenerationStatus(v1alpha1.GenerationStatus{ObservedGeneration: 1}).
				setConditions(
					metav1.Condition{
						Type:               conditions.TypeProgressing,
						Status:             metav1.ConditionTrue,
						Reason:             "ProgressingResource",
						ObservedGeneration: 1,
					},
					metav1.Condition{
						Type:               conditions.TypeReady,
						Status:             metav1.ConditionFalse,
						Reason:             "NotAvailable",
						ObservedGeneration: 1,
					},
				).getInstance(),
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			p := UpdateStandalonePipeline{}
			setClientInContext(ts.Context, ts.Client)
			setInstanceInContext(ts.Context, tc.givenInstance)
			tc.prepare(tc.givenInstance)

			// Act
			err := p.markInstanceAsProgressing(ts.Context)
			ts.Require().NoError(err)

			// Assert
			actualInstance := &v1alpha1.PostgresqlStandalone{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-first-instance",
					Namespace: "instance-first-namespace",
				},
			}
			ts.FetchResource(client.ObjectKeyFromObject(actualInstance), actualInstance)
			expectedStatus := tc.expectedInstance.Status
			ts.Assert().Equal(expectedStatus.ObservedGeneration, actualInstance.Status.ObservedGeneration)
			ts.Assert().Equal(expectedStatus.Conditions[0].Status, actualInstance.Status.Conditions[0].Status)
			ts.Assert().Equal(expectedStatus.Conditions[0].Type, actualInstance.Status.Conditions[0].Type)
			ts.Assert().Equal(expectedStatus.Conditions[0].Reason, actualInstance.Status.Conditions[0].Reason)
			ts.Assert().Equal(expectedStatus.Conditions[0].ObservedGeneration, actualInstance.Status.Conditions[0].ObservedGeneration)
			ts.Assert().Equal(expectedStatus.Conditions[1].Status, actualInstance.Status.Conditions[1].Status)
			ts.Assert().Equal(expectedStatus.Conditions[1].Type, actualInstance.Status.Conditions[1].Type)
			ts.Assert().Equal(expectedStatus.Conditions[1].Reason, actualInstance.Status.Conditions[1].Reason)
			ts.Assert().Equal(expectedStatus.Conditions[1].ObservedGeneration, actualInstance.Status.Conditions[1].ObservedGeneration)
		})
	}
}

func (ts *UpdateStandalonePipelineSuite) Test_MarkInstanceAsReady() {
	tests := map[string]struct {
		prepare          func(*v1alpha1.PostgresqlStandalone)
		givenInstance    *v1alpha1.PostgresqlStandalone
		expectedInstance *v1alpha1.PostgresqlStandalone
	}{
		"GivenInstanceWasUpdated_WhenSecondUpdateReconcileFinished_ThenExpectStatusReadyTrue": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				ts.EnsureNS("instance-second-namespace")
				ts.EnsureResources(instance)
			},
			givenInstance: newInstanceBuilder("my-second-instance", "instance-second-namespace").
				setConditions(
					metav1.Condition{
						Type:               conditions.TypeProgressing,
						Status:             metav1.ConditionTrue,
						Reason:             "ProgressingResource",
						ObservedGeneration: 1,
					},
				).
				getInstance(),
			expectedInstance: newInstanceBuilder("my-second-instance", "instance-second-namespace").
				setConditions(
					metav1.Condition{
						Type:               conditions.TypeReady,
						Status:             metav1.ConditionTrue,
						Reason:             "Available",
						ObservedGeneration: 1,
					},
				).getInstance(),
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			p := UpdateStandalonePipeline{}
			setInstanceInContext(ts.Context, tc.givenInstance)
			tc.prepare(tc.givenInstance)

			// Act
			err := p.markInstanceAsReady(ts.Context)
			ts.Require().NoError(err)

			// Assert
			actualInstance := &v1alpha1.PostgresqlStandalone{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-second-instance",
					Namespace: "instance-second-namespace",
				},
			}
			ts.FetchResource(client.ObjectKeyFromObject(actualInstance), actualInstance)
			expectedStatus := tc.expectedInstance.Status
			ts.Assert().Equal(1, len(actualInstance.Status.Conditions))
			ts.Assert().Equal(expectedStatus.Conditions[0].Status, actualInstance.Status.Conditions[0].Status)
			ts.Assert().Equal(expectedStatus.Conditions[0].Type, actualInstance.Status.Conditions[0].Type)
			ts.Assert().Equal(expectedStatus.Conditions[0].Reason, actualInstance.Status.Conditions[0].Reason)
			ts.Assert().Equal(expectedStatus.Conditions[0].ObservedGeneration, actualInstance.Status.Conditions[0].ObservedGeneration)
		})
	}
}
