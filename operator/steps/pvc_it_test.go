//go:build integration

package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"testing"
)

type PvcSuite struct {
	operatortest.Suite
}

func TestPvcSuite(t *testing.T) {
	suite.Run(t, new(PvcSuite))
}

func (ts *PvcSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	SetClientInContext(ts.Context, ts.Client)
	ts.RegisterScheme(storagev1.AddToScheme)
}

func (ts *PvcSuite) Test_EnsurePvcFn() {
	tests := map[string]struct {
		prepare                func()
		givenNamespace         string
		configuredAccessModes  []corev1.PersistentVolumeAccessMode
		configuredStorageClass *string
		givenStorageSize       *resource.Quantity
		expectedAccessModes    []corev1.PersistentVolumeAccessMode
		expectedStorageClass   *string
		expectedStorageSize    resource.Quantity
	}{
		"GivenNewPvc_WhenCreating_ThenCreateWithStorageClassAndAccessModes": {
			prepare:                func() {},
			givenNamespace:         "new-pvc",
			configuredAccessModes:  []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			configuredStorageClass: pointer.String("my-class"),
			givenStorageSize:       parseResource("1Gi"),
			expectedAccessModes:    []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			expectedStorageClass:   pointer.String("my-class"),
			expectedStorageSize:    *parseResource("1Gi"),
		},
		"GivenExistingPvc_WhenUpdating_ThenIgnoreStorageClassAndAccessModes": {
			prepare: func() {
				existingPvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: getPVCName(), Namespace: "existing-pvc"},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceStorage: *parseResource("1Gi")},
						},
						StorageClassName: pointer.String("existing-class")}}

				storageClass := &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "existing-class"},
					Provisioner:          "irrelevant-but-required",
					AllowVolumeExpansion: pointer.Bool(true)} // Required, otherwise we'll get an error while resizing.

				ts.EnsureResources(existingPvc, storageClass)
				// we need to set the status to "Bound", otherwise K8s doesn't accept resize.
				existingPvc.Status = corev1.PersistentVolumeClaimStatus{
					Phase:       corev1.ClaimBound,
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
					Capacity:    map[corev1.ResourceName]resource.Quantity{corev1.ResourceStorage: *parseResource("1Gi")}}
				ts.UpdateStatus(existingPvc)
			},
			givenNamespace:         "existing-pvc",
			configuredAccessModes:  []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			configuredStorageClass: pointer.String("my-class"),
			givenStorageSize:       parseResource("2Gi"),
			expectedAccessModes:    []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			expectedStorageClass:   pointer.String("existing-class"),
			expectedStorageSize:    *parseResource("2Gi"),
		},
	}

	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			config := newPostgresqlStandaloneOperatorConfig("config", "postgresql-system")
			config.Spec.Persistence.AccessModes = tc.configuredAccessModes
			config.Spec.Persistence.StorageClassName = tc.configuredStorageClass
			pipeline.StoreInContext(ts.Context, ConfigKey{}, config)
			pipeline.StoreInContext(ts.Context, DeploymentNamespaceKey{}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: tc.givenNamespace}})

			instance := NewInstanceBuilder("instance", "pvc-test").setDeploymentNamespace(tc.givenNamespace).getInstance()
			instance.Spec.Parameters.Resources.StorageCapacity = tc.givenStorageSize
			SetInstanceInContext(ts.Context, instance)
			ts.EnsureNS(tc.givenNamespace)
			tc.prepare()

			// Act
			err := EnsurePvcFn(labels.Set{"test": "label"})(ts.Context)
			ts.Require().NoError(err)

			// Assert
			result := &corev1.PersistentVolumeClaim{}
			ts.FetchResource(types.NamespacedName{Name: getPVCName(), Namespace: tc.givenNamespace}, result)

			ts.Assert().Equal(tc.expectedAccessModes, result.Spec.AccessModes, "access modes")
			ts.Assert().Equal(*tc.expectedStorageClass, *result.Spec.StorageClassName, "storage class")
			ts.Assert().True(result.Spec.Resources.Requests.Storage().Equal(tc.expectedStorageSize), "storage size")
			ts.Assert().Equal("label", result.Labels["test"], "label")
		})
	}
}
