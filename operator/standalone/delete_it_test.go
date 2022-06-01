//go:build integration

package standalone

import (
	"context"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	helmv1beta1 "github.com/crossplane-contrib/provider-helm/apis/release/v1beta1"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
)

type DeleteStandalonePipelineSuite struct {
	operatortest.Suite
}

func TestDeleteStandalonePipeline(t *testing.T) {
	suite.Run(t, new(DeleteStandalonePipelineSuite))
}

func (ts *DeleteStandalonePipelineSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	setClientInContext(ts.Context, ts.Client)
	ts.RegisterScheme(helmv1beta1.SchemeBuilder.AddToScheme)
}

func (ts *DeleteStandalonePipelineSuite) Test_DeleteHelmRelease() {
	tests := map[string]struct {
		prepare                    func(releaseNameString string)
		givenReleaseName           string
		expectedHelmReleaseDeleted bool
		expectedError              string
	}{
		"GivenNonExistingHelmRelease_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare:                    func(releaseName string) {},
			givenReleaseName:           "postgresql-release",
			expectedHelmReleaseDeleted: true,
			expectedError:              "",
		},
		"GivenAnExistingHelmRelease_WhenHelmReleaseStillExists_ThenExpectReconciliation": {
			prepare: func(releaseName string) {
				release := newPostgresqlHelmRelease(releaseName)
				ts.EnsureResources(release)
			},
			givenReleaseName:           "postgresql-release",
			expectedHelmReleaseDeleted: false,
			expectedError:              "",
		},
		"GivenAnExistingHelmRelease_WhenDeletingGeneratesAnError_ThenReturnError": {
			prepare: func(releaseName string) {},
			// we purposefully set namespace to empty so that we generate an error so that we avoid mocking.
			givenReleaseName:           "",
			expectedHelmReleaseDeleted: false,
			expectedError:              "resource name may not be empty",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			d := &DeleteStandalonePipeline{
				client: ts.Client,
				instance: newBuilderInstance("instance", "namespace").
					setDeploymentNamespace(tc.givenReleaseName).
					get(),
				helmReleaseDeleted: false,
			}
			tc.prepare(tc.givenReleaseName)
			err := d.deleteHelmRelease(ts.Context)

			// Assert
			ts.Assert().Equal(tc.expectedHelmReleaseDeleted, d.helmReleaseDeleted)
			if tc.expectedError != "" {
				ts.Assert().EqualError(err, tc.expectedError)
				return
			}
			ts.Assert().NoError(err)
			resultRelease := &helmv1beta1.Release{}
			err = ts.Client.Get(
				ts.Context,
				client.ObjectKey{Name: tc.givenReleaseName},
				resultRelease,
			)
			ts.AssertResourceNotExists(resultRelease.GetDeletionTimestamp(), err)
		})
	}
}

func (ts *DeleteStandalonePipelineSuite) Test_DeleteNamespace() {
	tests := map[string]struct {
		prepare        func(namespace string)
		givenNamespace string
		expectedError  string
	}{
		"GivenNonExistingNamespace_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare:        func(namespace string) {},
			givenNamespace: "non-existing-namespace",
			expectedError:  "",
		},
		"GivenExistingNamespace_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare:        func(namespace string) { ts.EnsureNS(namespace) },
			givenNamespace: "existing-namespace",
			expectedError:  "",
		},
		"GivenExistingNamespace_WhenDeletingGeneratesAnError_ThenReturnError": {
			prepare: func(namespace string) { ts.EnsureNS("an-existing-namespace") },
			// we purposefully set namespace to empty so that we generate an error so that we avoid mocking.
			givenNamespace: "",
			expectedError:  "resource name may not be empty",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			d := &DeleteStandalonePipeline{
				client: ts.Client,
				instance: newBuilderInstance("instance", "namespace").
					setDeploymentNamespace(tc.givenNamespace).
					get(),
				helmReleaseDeleted: false,
			}
			tc.prepare(tc.givenNamespace)
			err := d.deleteNamespace(ts.Context)

			// Assert
			if tc.expectedError != "" {
				ts.Assert().EqualError(err, tc.expectedError)
				return
			}
			ts.Assert().NoError(err)
			resultNs := &corev1.Namespace{}
			err = ts.Client.Get(
				ts.Context,
				types.NamespacedName{Name: tc.givenNamespace},
				resultNs,
			)
			ts.AssertResourceNotExists(resultNs.GetDeletionTimestamp(), err)
		})
	}
}

func (ts *DeleteStandalonePipelineSuite) Test_DeleteConnectionSecret() {
	tests := map[string]struct {
		prepare        func(name, namespace string)
		givenNamespace string
		givenSecret    string
		expectedError  string
	}{
		"GivenNonExistingSecret_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare:        func(name, namespace string) {},
			givenNamespace: "test-namespace",
			givenSecret:    "non-existing-secret",
			expectedError:  "",
		},
		"GivenExistingSecret_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare: func(name, namespace string) {
				ts.EnsureNS(namespace)
				ts.EnsureResources(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}})
			},
			givenNamespace: "test-namespace",
			givenSecret:    "existing-secret",
			expectedError:  "",
		},
		"GivenExistingSecret_WhenDeletingGeneratesAnError_ThenReturnError": {
			prepare: func(name, namespace string) {
				ts.EnsureNS(namespace)
				ts.EnsureResources(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: namespace}})
			},
			givenNamespace: "test-namespace",
			// we purposefully set namespace to empty so that we generate an error so that we avoid mocking.
			givenSecret:   "",
			expectedError: "resource name may not be empty",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			d := &DeleteStandalonePipeline{
				client: ts.Client,
				instance: newBuilderInstance("instance", tc.givenNamespace).
					setConnectionSecret(tc.givenSecret).
					get(),
				helmReleaseDeleted: false,
			}
			tc.prepare(tc.givenSecret, tc.givenNamespace)
			err := d.deleteConnectionSecret(ts.Context)

			// Assert
			if tc.expectedError != "" {
				ts.Assert().EqualError(err, tc.expectedError)
				return
			}
			ts.Assert().NoError(err)
			resultSecret := &corev1.Secret{}
			err = ts.Client.Get(
				ts.Context,
				types.NamespacedName{Name: tc.givenSecret, Namespace: tc.givenNamespace},
				resultSecret,
			)
			ts.AssertResourceNotExists(resultSecret.GetDeletionTimestamp(), err)
		})
	}
}

func (ts *DeleteStandalonePipelineSuite) Test_RemoveFinalizer() {
	tests := map[string]struct {
		prepare        func(instance *v1alpha1.PostgresqlStandalone)
		givenInstance  string
		givenNamespace string
	}{
		"GivenAnInstance_WhenDeletingFinalizer_ThenExpectInstanceWithNoFinalizer": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				ts.EnsureNS("remove-finalizer")
				ts.EnsureResources(instance)
			},

			givenInstance:  "instance",
			givenNamespace: "remove-finalizer",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			instance := newBuilderInstance(tc.givenInstance, tc.givenNamespace).setFinalizers(finalizer).get()
			d := &DeleteStandalonePipeline{
				client:             ts.Client,
				instance:           instance,
				helmReleaseDeleted: false,
			}
			tc.prepare(instance)
			err := d.removeFinalizer(ts.Context)
			// Assert
			ts.Require().NoError(err)
			releaseResult := &helmv1beta1.Release{}
			err = ts.Client.Get(
				ts.Context,
				types.NamespacedName{Name: tc.givenInstance, Namespace: tc.givenNamespace},
				releaseResult,
			)
			ts.Assert().Empty(releaseResult.Finalizers)
		})
	}
}

func newPostgresqlHelmRelease(name string) *helmv1beta1.Release {
	return &helmv1beta1.Release{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

// AssertResourceNotExists checks if the given resource is not existing or is existing with a deletion timestamp.
// Test fails if the resource exists or there's another error.
func (ts *DeleteStandalonePipelineSuite) AssertResourceNotExists(deletionTime *metav1.Time, err error) {
	if err != nil {
		ts.Require().True(apierrors.IsNotFound(err))
	} else {
		ts.Require().False(deletionTime.IsZero())
	}
}

type PostgresqlStandaloneBuilder struct {
	*v1alpha1.PostgresqlStandalone
}

func newBuilderInstance(name, namespace string) *PostgresqlStandaloneBuilder {
	return &PostgresqlStandaloneBuilder{newInstance(name, namespace)}
}

func (b *PostgresqlStandaloneBuilder) setDeploymentNamespace(namespace string) *PostgresqlStandaloneBuilder {
	b.Status = v1alpha1.PostgresqlStandaloneStatus{
		PostgresqlStandaloneObservation: v1alpha1.PostgresqlStandaloneObservation{
			HelmChart: &v1alpha1.ChartMetaStatus{
				DeploymentNamespace: namespace,
			},
		},
	}
	return b
}

func (b *PostgresqlStandaloneBuilder) setConnectionSecret(secret string) *PostgresqlStandaloneBuilder {
	b.Spec = v1alpha1.PostgresqlStandaloneSpec{
		ConnectableInstance: v1alpha1.ConnectableInstance{
			WriteConnectionSecretToRef: v1alpha1.ConnectionSecretRef{
				Name: secret,
			},
		},
	}
	return b
}

func (b *PostgresqlStandaloneBuilder) setFinalizers(finalizers ...string) *PostgresqlStandaloneBuilder {
	b.Finalizers = finalizers
	return b
}

func (b *PostgresqlStandaloneBuilder) get() *v1alpha1.PostgresqlStandalone {
	return b.PostgresqlStandalone
}
