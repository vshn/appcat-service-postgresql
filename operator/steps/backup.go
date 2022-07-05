package steps

import (
	"context"
	pipeline "github.com/ccremer/go-command-pipeline"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	"github.com/vshn/appcat-service-postgresql/apis/postgresql/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// IsBackupEnabledP returns a predicate that returns true if backups are enabled in the spec.
func IsBackupEnabledP() pipeline.Predicate {
	return func(ctx context.Context) bool {
		instance := GetInstanceFromContext(ctx)

		return instance.Spec.Backup.Enabled
	}
}

// EnsureK8upScheduleFn creates the K8up schedule object.
func EnsureK8upScheduleFn(labelSet labels.Set) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		kube := GetClientFromContext(ctx)
		instance := GetInstanceFromContext(ctx)
		config := GetConfigFromContext(ctx)
		bucketSecret := getFromContextOrPanic(ctx, BucketSecretKey{}).(*corev1.Secret)

		schedule := newK8upSchedule(instance)

		_, err := controllerutil.CreateOrUpdate(ctx, kube, schedule, func() error {
			schedule.Labels = labels.Merge(schedule.Labels, labelSet)
			schedule.Spec = k8upv1.ScheduleSpec{
				Backup: &k8upv1.BackupSchedule{
					BackupSpec:     k8upv1.BackupSpec{},
					ScheduleCommon: &k8upv1.ScheduleCommon{Schedule: "@daily-random"},
				},
				Archive: &k8upv1.ArchiveSchedule{
					ScheduleCommon: &k8upv1.ScheduleCommon{Schedule: "@weekly-random"},
				},
				Check: &k8upv1.CheckSchedule{
					ScheduleCommon: &k8upv1.ScheduleCommon{Schedule: "@weekly-random"},
				},
				Prune: &k8upv1.PruneSchedule{
					ScheduleCommon: &k8upv1.ScheduleCommon{Schedule: "@weekly-random"},
				},
				Backend: &k8upv1.Backend{
					RepoPasswordSecretRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: getResticRepositorySecretName()},
						Key:                  "repository",
					},
					S3: &k8upv1.S3Spec{
						Endpoint:                 string(bucketSecret.Data[config.Spec.BackupConfigSpec.S3BucketSecret.EndpointRef.Key]),
						Bucket:                   string(bucketSecret.Data[config.Spec.BackupConfigSpec.S3BucketSecret.BucketRef.Key]),
						AccessKeyIDSecretRef:     &config.Spec.BackupConfigSpec.S3BucketSecret.AccessKeyRef,
						SecretAccessKeySecretRef: &config.Spec.BackupConfigSpec.S3BucketSecret.SecretKeyRef,
					},
				},
				FailedJobsHistoryLimit:     pointer.Int(2),
				SuccessfulJobsHistoryLimit: pointer.Int(2),
			}
			return nil
		})
		return err
	}
}

// DeleteK8upScheduleFn deletes the K8up schedule associated to the instance.
// If the resource doesn't exist, it returns nil (no-op).
func DeleteK8upScheduleFn() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		instance := GetInstanceFromContext(ctx)
		kube := GetClientFromContext(ctx)

		if instance.Status.HelmChart == nil || instance.Status.HelmChart.DeploymentNamespace == "" {
			// deployment namespace is unknown, we assume it has not been created
			return nil
		}
		schedule := newK8upSchedule(instance)
		err := kube.Delete(ctx, schedule)
		return client.IgnoreNotFound(err)
	}
}

func newK8upSchedule(instance *v1alpha1.PostgresqlStandalone) *k8upv1.Schedule {
	return &k8upv1.Schedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql",
			Namespace: instance.Status.HelmChart.DeploymentNamespace,
		},
	}
}
