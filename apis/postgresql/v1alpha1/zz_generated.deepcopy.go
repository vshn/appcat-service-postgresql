//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupConfigSpec) DeepCopyInto(out *BackupConfigSpec) {
	*out = *in
	in.S3BucketSecret.DeepCopyInto(&out.S3BucketSecret)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupConfigSpec.
func (in *BackupConfigSpec) DeepCopy() *BackupConfigSpec {
	if in == nil {
		return nil
	}
	out := new(BackupConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupEnabledInstance) DeepCopyInto(out *BackupEnabledInstance) {
	*out = *in
	out.Backup = in.Backup
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupEnabledInstance.
func (in *BackupEnabledInstance) DeepCopy() *BackupEnabledInstance {
	if in == nil {
		return nil
	}
	out := new(BackupEnabledInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupSpec) DeepCopyInto(out *BackupSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupSpec.
func (in *BackupSpec) DeepCopy() *BackupSpec {
	if in == nil {
		return nil
	}
	out := new(BackupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChartMeta) DeepCopyInto(out *ChartMeta) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChartMeta.
func (in *ChartMeta) DeepCopy() *ChartMeta {
	if in == nil {
		return nil
	}
	out := new(ChartMeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChartMetaStatus) DeepCopyInto(out *ChartMetaStatus) {
	*out = *in
	out.ChartMeta = in.ChartMeta
	in.ModifiedTime.DeepCopyInto(&out.ModifiedTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChartMetaStatus.
func (in *ChartMetaStatus) DeepCopy() *ChartMetaStatus {
	if in == nil {
		return nil
	}
	out := new(ChartMetaStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComputeResources) DeepCopyInto(out *ComputeResources) {
	*out = *in
	if in.MemoryLimit != nil {
		in, out := &in.MemoryLimit, &out.MemoryLimit
		x := (*in).DeepCopy()
		*out = &x
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComputeResources.
func (in *ComputeResources) DeepCopy() *ComputeResources {
	if in == nil {
		return nil
	}
	out := new(ComputeResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectableInstance) DeepCopyInto(out *ConnectableInstance) {
	*out = *in
	out.WriteConnectionSecretToRef = in.WriteConnectionSecretToRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectableInstance.
func (in *ConnectableInstance) DeepCopy() *ConnectableInstance {
	if in == nil {
		return nil
	}
	out := new(ConnectableInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectionSecretRef) DeepCopyInto(out *ConnectionSecretRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectionSecretRef.
func (in *ConnectionSecretRef) DeepCopy() *ConnectionSecretRef {
	if in == nil {
		return nil
	}
	out := new(ConnectionSecretRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenerationStatus) DeepCopyInto(out *GenerationStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenerationStatus.
func (in *GenerationStatus) DeepCopy() *GenerationStatus {
	if in == nil {
		return nil
	}
	out := new(GenerationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmReleaseConfig) DeepCopyInto(out *HelmReleaseConfig) {
	*out = *in
	out.Chart = in.Chart
	in.Values.DeepCopyInto(&out.Values)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmReleaseConfig.
func (in *HelmReleaseConfig) DeepCopy() *HelmReleaseConfig {
	if in == nil {
		return nil
	}
	out := new(HelmReleaseConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Persistence) DeepCopyInto(out *Persistence) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Persistence.
func (in *Persistence) DeepCopy() *Persistence {
	if in == nil {
		return nil
	}
	out := new(Persistence)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandalone) DeepCopyInto(out *PostgresqlStandalone) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandalone.
func (in *PostgresqlStandalone) DeepCopy() *PostgresqlStandalone {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandalone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PostgresqlStandalone) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneConfigStatus) DeepCopyInto(out *PostgresqlStandaloneConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneConfigStatus.
func (in *PostgresqlStandaloneConfigStatus) DeepCopy() *PostgresqlStandaloneConfigStatus {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneList) DeepCopyInto(out *PostgresqlStandaloneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PostgresqlStandalone, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneList.
func (in *PostgresqlStandaloneList) DeepCopy() *PostgresqlStandaloneList {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PostgresqlStandaloneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneObservation) DeepCopyInto(out *PostgresqlStandaloneObservation) {
	*out = *in
	if in.HelmChart != nil {
		in, out := &in.HelmChart, &out.HelmChart
		*out = new(ChartMetaStatus)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneObservation.
func (in *PostgresqlStandaloneObservation) DeepCopy() *PostgresqlStandaloneObservation {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneOperatorConfig) DeepCopyInto(out *PostgresqlStandaloneOperatorConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneOperatorConfig.
func (in *PostgresqlStandaloneOperatorConfig) DeepCopy() *PostgresqlStandaloneOperatorConfig {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneOperatorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PostgresqlStandaloneOperatorConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneOperatorConfigList) DeepCopyInto(out *PostgresqlStandaloneOperatorConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PostgresqlStandaloneOperatorConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneOperatorConfigList.
func (in *PostgresqlStandaloneOperatorConfigList) DeepCopy() *PostgresqlStandaloneOperatorConfigList {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneOperatorConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PostgresqlStandaloneOperatorConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneOperatorConfigSpec) DeepCopyInto(out *PostgresqlStandaloneOperatorConfigSpec) {
	*out = *in
	in.ResourceMinima.DeepCopyInto(&out.ResourceMinima)
	in.ResourceMaxima.DeepCopyInto(&out.ResourceMaxima)
	if in.HelmReleaseTemplate != nil {
		in, out := &in.HelmReleaseTemplate, &out.HelmReleaseTemplate
		*out = new(HelmReleaseConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.HelmReleases != nil {
		in, out := &in.HelmReleases, &out.HelmReleases
		*out = make([]HelmReleaseConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.BackupConfigSpec.DeepCopyInto(&out.BackupConfigSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneOperatorConfigSpec.
func (in *PostgresqlStandaloneOperatorConfigSpec) DeepCopy() *PostgresqlStandaloneOperatorConfigSpec {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneOperatorConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneParameters) DeepCopyInto(out *PostgresqlStandaloneParameters) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneParameters.
func (in *PostgresqlStandaloneParameters) DeepCopy() *PostgresqlStandaloneParameters {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneSpec) DeepCopyInto(out *PostgresqlStandaloneSpec) {
	*out = *in
	out.ConnectableInstance = in.ConnectableInstance
	out.BackupEnabledInstance = in.BackupEnabledInstance
	in.Parameters.DeepCopyInto(&out.Parameters)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneSpec.
func (in *PostgresqlStandaloneSpec) DeepCopy() *PostgresqlStandaloneSpec {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneStatus) DeepCopyInto(out *PostgresqlStandaloneStatus) {
	*out = *in
	out.GenerationStatus = in.GenerationStatus
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.PostgresqlStandaloneObservation.DeepCopyInto(&out.PostgresqlStandaloneObservation)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneStatus.
func (in *PostgresqlStandaloneStatus) DeepCopy() *PostgresqlStandaloneStatus {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resources) DeepCopyInto(out *Resources) {
	*out = *in
	in.ComputeResources.DeepCopyInto(&out.ComputeResources)
	in.StorageResources.DeepCopyInto(&out.StorageResources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Resources.
func (in *Resources) DeepCopy() *Resources {
	if in == nil {
		return nil
	}
	out := new(Resources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageResources) DeepCopyInto(out *StorageResources) {
	*out = *in
	if in.StorageCapacity != nil {
		in, out := &in.StorageCapacity, &out.StorageCapacity
		x := (*in).DeepCopy()
		*out = &x
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageResources.
func (in *StorageResources) DeepCopy() *StorageResources {
	if in == nil {
		return nil
	}
	out := new(StorageResources)
	in.DeepCopyInto(out)
	return out
}
