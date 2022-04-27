//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backup) DeepCopyInto(out *Backup) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backup.
func (in *Backup) DeepCopy() *Backup {
	if in == nil {
		return nil
	}
	out := new(Backup)
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
	if in.ModifiedAt != nil {
		in, out := &in.ModifiedAt, &out.ModifiedAt
		*out = (*in).DeepCopy()
	}
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
func (in *DeferrableMaintenance) DeepCopyInto(out *DeferrableMaintenance) {
	*out = *in
	in.UpdatePolicy.DeepCopyInto(&out.UpdatePolicy)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeferrableMaintenance.
func (in *DeferrableMaintenance) DeepCopy() *DeferrableMaintenance {
	if in == nil {
		return nil
	}
	out := new(DeferrableMaintenance)
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
func (in *MaintenanceWindow) DeepCopyInto(out *MaintenanceWindow) {
	*out = *in
	in.Start.DeepCopyInto(&out.Start)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaintenanceWindow.
func (in *MaintenanceWindow) DeepCopy() *MaintenanceWindow {
	if in == nil {
		return nil
	}
	out := new(MaintenanceWindow)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaintenanceWindowSelector) DeepCopyInto(out *MaintenanceWindowSelector) {
	*out = *in
	if in.Hour != nil {
		in, out := &in.Hour, &out.Hour
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaintenanceWindowSelector.
func (in *MaintenanceWindowSelector) DeepCopy() *MaintenanceWindowSelector {
	if in == nil {
		return nil
	}
	out := new(MaintenanceWindowSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Monitoring) DeepCopyInto(out *Monitoring) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Monitoring.
func (in *Monitoring) DeepCopy() *Monitoring {
	if in == nil {
		return nil
	}
	out := new(Monitoring)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonitoringEnabledInstance) DeepCopyInto(out *MonitoringEnabledInstance) {
	*out = *in
	out.Monitoring = in.Monitoring
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonitoringEnabledInstance.
func (in *MonitoringEnabledInstance) DeepCopy() *MonitoringEnabledInstance {
	if in == nil {
		return nil
	}
	out := new(MonitoringEnabledInstance)
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
func (in *PostgresqlStandaloneConfig) DeepCopyInto(out *PostgresqlStandaloneConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneConfig.
func (in *PostgresqlStandaloneConfig) DeepCopy() *PostgresqlStandaloneConfig {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PostgresqlStandaloneConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneConfigList) DeepCopyInto(out *PostgresqlStandaloneConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PostgresqlStandaloneConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneConfigList.
func (in *PostgresqlStandaloneConfigList) DeepCopy() *PostgresqlStandaloneConfigList {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PostgresqlStandaloneConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresqlStandaloneConfigSpec) DeepCopyInto(out *PostgresqlStandaloneConfigSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresqlStandaloneConfigSpec.
func (in *PostgresqlStandaloneConfigSpec) DeepCopy() *PostgresqlStandaloneConfigSpec {
	if in == nil {
		return nil
	}
	out := new(PostgresqlStandaloneConfigSpec)
	in.DeepCopyInto(out)
	return out
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
func (in *PostgresqlStandaloneSpec) DeepCopyInto(out *PostgresqlStandaloneSpec) {
	*out = *in
	out.BackupEnabledInstance = in.BackupEnabledInstance
	out.MonitoringEnabledInstance = in.MonitoringEnabledInstance
	in.DeferrableMaintenance.DeepCopyInto(&out.DeferrableMaintenance)
	out.Resources = in.Resources
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
	out.ComputeResources = in.ComputeResources
	out.StorageResources = in.StorageResources
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpdatePolicy) DeepCopyInto(out *UpdatePolicy) {
	*out = *in
	out.Version = in.Version
	in.MaintenanceWindow.DeepCopyInto(&out.MaintenanceWindow)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpdatePolicy.
func (in *UpdatePolicy) DeepCopy() *UpdatePolicy {
	if in == nil {
		return nil
	}
	out := new(UpdatePolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionSelector) DeepCopyInto(out *VersionSelector) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionSelector.
func (in *VersionSelector) DeepCopy() *VersionSelector {
	if in == nil {
		return nil
	}
	out := new(VersionSelector)
	in.DeepCopyInto(out)
	return out
}
