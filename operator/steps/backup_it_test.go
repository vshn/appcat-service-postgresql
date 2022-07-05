package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"github.com/stretchr/testify/suite"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	"github.com/vshn/appcat-service-postgresql/operator/operatortest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type K8upBackupSuite struct {
	operatortest.Suite
}

func TestK8upBackupSuite(t *testing.T) {
	suite.Run(t, new(K8upBackupSuite))
}

func (ts *K8upBackupSuite) BeforeTest(suiteName, testName string) {
	ts.Context = pipeline.MutableContext(context.Background())
	SetClientInContext(ts.Context, ts.Client)
	ts.RegisterScheme(k8upv1.SchemeBuilder.AddToScheme)
}

func (ts *K8upBackupSuite) Test_EnsureK8upSchedule() {
	type testCase struct {
		prepare       func(testCase)
		givenInstance *v1alpha1.PostgresqlStandalone
	}
	tests := map[string]testCase{
		"GivenScheduleDoesNotExist_WhenCreatingSchedule_ThenExpectNewSchedule": {
			prepare: func(tc testCase) {
				ts.EnsureNS("new-schedule")
			},
			givenInstance: NewInstanceBuilder("instance", "postgresql-instance").
				setDeploymentNamespace("new-schedule").
				setBackupEnabled(true).
				getInstance(),
		},
		"GiveExistingSchedule_WhenUpdatingSchedule_ThenRevertSpecOfOldSchedule": {
			prepare: func(tc testCase) {
				ts.EnsureNS("existing-schedule")
				existingSchedule := newK8upSchedule(tc.givenInstance)
				existingSchedule.Spec.Archive = &k8upv1.ArchiveSchedule{
					ArchiveSpec: k8upv1.ArchiveSpec{
						RestoreSpec: &k8upv1.RestoreSpec{Tags: []string{"tag"}}, // this should be removed
					},
				}
				ts.EnsureResources(existingSchedule)
			},
			givenInstance: NewInstanceBuilder("instance", "postgresql-instance").
				setDeploymentNamespace("existing-schedule").
				setBackupEnabled(true).
				getInstance(),
		},
	}
	for name, tc := range tests {
		ts.Run(name, func() {
			// Arrange
			SetClientInContext(ts.Context, ts.Client)
			SetInstanceInContext(ts.Context, tc.givenInstance)
			config := &v1alpha1.PostgresqlStandaloneOperatorConfig{
				Spec: v1alpha1.PostgresqlStandaloneOperatorConfigSpec{
					BackupConfigSpec: v1alpha1.BackupConfigSpec{
						S3BucketSecret: v1alpha1.S3BucketConfigSpec{
							AccessKeyRef: corev1.SecretKeySelector{Key: "accessKey", LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
							SecretKeyRef: corev1.SecretKeySelector{Key: "secretKey", LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
							BucketRef:    corev1.SecretKeySelector{Key: "bucket", LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
							EndpointRef:  corev1.SecretKeySelector{Key: "endpoint", LocalObjectReference: corev1.LocalObjectReference{Name: "s3-credentials"}},
						},
					},
				},
			}
			pipeline.StoreInContext(ts.Context, ConfigKey{}, config)
			bucketSecret := &corev1.Secret{
				Data: map[string][]byte{
					"accessKey": []byte("access"),
					"secretKey": []byte("secret"),
					"bucket":    []byte("k8up-bucket"),
					"endpoint":  []byte("http://minio:9000"),
				},
			}
			pipeline.StoreInContext(ts.Context, BucketSecretKey{}, bucketSecret)

			tc.prepare(tc)

			// Act
			err := EnsureK8upScheduleFn(labels.Set{})(ts.Context)
			ts.Require().NoError(err)

			// Assert
			result := &k8upv1.Schedule{}
			err = ts.Client.Get(ts.Context, client.ObjectKeyFromObject(newK8upSchedule(tc.givenInstance)), result)

			ts.Assert().Nil(result.Spec.Archive.RestoreSpec, "restore spec")
			ts.Assert().Equal("@daily-random", result.Spec.Backup.ScheduleCommon.Schedule.String(), "backup schedule")
			ts.Assert().Equal("@weekly-random", result.Spec.Archive.ScheduleCommon.Schedule.String(), "archive schedule")
			ts.Assert().Equal("@weekly-random", result.Spec.Prune.ScheduleCommon.Schedule.String(), "prune schedule")
			ts.Assert().Equal("@weekly-random", result.Spec.Check.ScheduleCommon.Schedule.String(), "check schedule")
			ts.Assert().Equal("repository", result.Spec.Backend.RepoPasswordSecretRef.Key, "repo encryption key ref")
			ts.Assert().Equal(2, *result.Spec.FailedJobsHistoryLimit, "failed jobs history limit")
			ts.Assert().Equal(2, *result.Spec.SuccessfulJobsHistoryLimit, "successful jobs history limit")
			ts.Assert().Equal("postgresql-restic", result.Spec.Backend.RepoPasswordSecretRef.LocalObjectReference.Name, "repo encryption name ref")
			ts.Assert().Equal("http://minio:9000", result.Spec.Backend.S3.Endpoint, "s3 endpoint")
			ts.Assert().Equal("k8up-bucket", result.Spec.Backend.S3.Bucket, "s3 bucket name")
			ts.Assert().Equal("accessKey", result.Spec.Backend.S3.AccessKeyIDSecretRef.Key)
			ts.Assert().Equal("s3-credentials", result.Spec.Backend.S3.AccessKeyIDSecretRef.Name)
			ts.Assert().Equal("secretKey", result.Spec.Backend.S3.SecretAccessKeySecretRef.Key)
			ts.Assert().Equal("s3-credentials", result.Spec.Backend.S3.SecretAccessKeySecretRef.Name)
		})
	}
}
