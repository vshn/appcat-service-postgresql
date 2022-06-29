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
	}{
		"GivenNonExistingHelmRelease_WhenDeleting_ThenExpectNoError": {
			prepare:                    func(releaseName string) {},
			givenReleaseName:           "postgresql-release",
			expectedHelmReleaseDeleted: true,
		},
		"GivenAnExistingHelmRelease_WhenHelmReleaseStillExists_ThenExpectNoError": {
			prepare: func(releaseName string) {
				release := newPostgresqlHelmRelease(releaseName)
				ts.EnsureResources(release)
			},
			givenReleaseName:           "postgresql-release",
			expectedHelmReleaseDeleted: false,
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			d := &DeleteStandalonePipeline{
				client: ts.Client,
				instance: newInstanceBuilder("instance", "namespace").
					setDeploymentNamespace(tc.givenReleaseName).
					get(),
				helmReleaseDeleted: false,
			}
			tc.prepare(tc.givenReleaseName)

			// Act
			err := d.deleteHelmRelease(ts.Context)
			ts.Require().NoError(err)

			// Assert
			ts.Assert().Equal(tc.expectedHelmReleaseDeleted, d.helmReleaseDeleted)
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
	}{
		"GivenNonExistingNamespace_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare:        func(namespace string) {},
			givenNamespace: "non-existing-namespace",
		},
		"GivenExistingNamespace_WhenDeleting_ThenExpectNoFurtherAction": {
			prepare:        func(namespace string) { ts.EnsureNS(namespace) },
			givenNamespace: "existing-namespace",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			d := &DeleteStandalonePipeline{
				client: ts.Client,
				instance: newInstanceBuilder("instance", "namespace").
					setDeploymentNamespace(tc.givenNamespace).
					get(),
				helmReleaseDeleted: false,
			}
			tc.prepare(tc.givenNamespace)

			// Act
			err := d.deleteNamespace(ts.Context)
			ts.Require().NoError(err)

			// Assert
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
		"GivenNonExistingSecret_WhenDeleting_ThenExpectNoError": {
			prepare:        func(name, namespace string) {},
			givenNamespace: "test-namespace",
			givenSecret:    "non-existing-secret",
			expectedError:  "",
		},
		"GivenExistingSecret_WhenDeleting_ThenExpectNoError": {
			prepare: func(name, namespace string) {
				ts.EnsureNS(namespace)
				ts.EnsureResources(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}})
			},
			givenNamespace: "test-namespace",
			givenSecret:    "existing-secret",
			expectedError:  "",
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			d := &DeleteStandalonePipeline{
				client: ts.Client,
				instance: newInstanceBuilder("instance", tc.givenNamespace).
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
		assert         func(previousInstance, result *v1alpha1.PostgresqlStandalone)
	}{
		"GivenInstanceWithFinalizer_WhenDeletingFinalizer_ThenExpectInstanceUpdatedWithRemovedFinalizer": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				instance.Finalizers = []string{finalizer}
				ts.EnsureNS("remove-finalizer")
				ts.EnsureResources(instance)
				ts.Assert().NotEmpty(instance.Finalizers)
			},

			givenInstance:  "has-finalizer",
			givenNamespace: "remove-finalizer",
			assert: func(previousInstance, result *v1alpha1.PostgresqlStandalone) {
				ts.Assert().Empty(result.Finalizers)
				ts.Assert().NotEqual(previousInstance.ResourceVersion, result.ResourceVersion, "resource version should change")
			},
		},
		"GivenInstanceWithoutFinalizer_WhenDeletingFinalizer_ThenExpectInstanceUnchanged": {
			prepare: func(instance *v1alpha1.PostgresqlStandalone) {
				ts.EnsureNS("remove-finalizer")
				ts.EnsureResources(instance)
			},

			givenInstance:  "no-finalizer",
			givenNamespace: "remove-finalizer",
			assert: func(previousInstance, result *v1alpha1.PostgresqlStandalone) {
				ts.Assert().Empty(result.Finalizers)
				ts.Assert().Equal(previousInstance.ResourceVersion, result.ResourceVersion, "resource version should be equal")
			},
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			instance := newInstanceBuilder(tc.givenInstance, tc.givenNamespace).get()
			d := &DeleteStandalonePipeline{
				client:             ts.Client,
				instance:           instance,
				helmReleaseDeleted: false,
			}
			tc.prepare(instance)
			previousVersion := instance.DeepCopy()

			// Act
			err := d.removeFinalizer(ts.Context)
			ts.Require().NoError(err)

			// Assert
			result := &v1alpha1.PostgresqlStandalone{}
			ts.FetchResource(client.ObjectKeyFromObject(instance), result)
			tc.assert(previousVersion, result)
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
