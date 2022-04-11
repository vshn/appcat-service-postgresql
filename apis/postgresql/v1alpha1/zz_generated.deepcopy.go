//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
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
func (in *DelayableMaintenance) DeepCopyInto(out *DelayableMaintenance) {
	*out = *in
	out.UpdatePolicy = in.UpdatePolicy
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DelayableMaintenance.
func (in *DelayableMaintenance) DeepCopy() *DelayableMaintenance {
	if in == nil {
		return nil
	}
	out := new(DelayableMaintenance)
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
	out.Start = in.Start
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
	if in.Chart != nil {
		in, out := &in.Chart, &out.Chart
		*out = new(ChartMeta)
		**out = **in
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
func (in *PostgresqlStandaloneParameters) DeepCopyInto(out *PostgresqlStandaloneParameters) {
	*out = *in
	if in.Chart != nil {
		in, out := &in.Chart, &out.Chart
		*out = new(ChartMeta)
		**out = **in
	}
	out.BackupEnabledInstance = in.BackupEnabledInstance
	out.MonitoringEnabledInstance = in.MonitoringEnabledInstance
	out.DelayableMaintenance = in.DelayableMaintenance
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
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
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
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.GenerationStatus = in.GenerationStatus
	in.AtProvider.DeepCopyInto(&out.AtProvider)
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
func (in *UpdatePolicy) DeepCopyInto(out *UpdatePolicy) {
	*out = *in
	out.Version = in.Version
	out.MaintenanceWindow = in.MaintenanceWindow
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
