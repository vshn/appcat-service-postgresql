package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reasons that give more context to conditions.
const (
	ReasonMaintenanceProgressing = "MaintenanceProgressing"
	ReasonMaintenanceSuccess     = "MaintenanceFinishedSuccessfully"
	ReasonMaintenanceFailure     = "MaintenanceFinishedWithError"
	ReasonReady                  = "Available"
	ReasonCreating               = "CreatingResources"
	ReasonProvisioning           = "Progressing"
)

const (
	// TypeInMaintenance identifies a condition related to maintenance.
	TypeInMaintenance = "InMaintenance"
	// TypeReady indicates that an instance is ready to serve.
	TypeReady = "Ready"
	// TypeProvisioning indicates that an instance is being provisioned.
	TypeProvisioning = "Provisioning"
	// TypeCreating indicates that an instance is being created for the first time.
	TypeCreating = "Creating"
)

// Ready creates a condition with TypeReady, ReasonReady and empty message.
func Ready() metav1.Condition {
	return metav1.Condition{
		Type:               TypeReady,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonReady,
	}
}

// Provisioning creates a condition with TypeProvisioning, ReasonProvisioning and empty message.
func Provisioning() metav1.Condition {
	return metav1.Condition{
		Type:               TypeProvisioning,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonProvisioning,
	}
}

// Creating creates a condition with TypeCreating, ReasonCreating and empty message.
func Creating() metav1.Condition {
	return metav1.Condition{
		Type:               TypeCreating,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonCreating,
	}
}

// InMaintenance creates an active condition with TypeReady, ReasonReady and empty message.
func InMaintenance() metav1.Condition {
	return metav1.Condition{
		Type:               TypeInMaintenance,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonMaintenanceProgressing,
	}
}

// MaintenanceSuccess creates an inactive condition with TypeInMaintenance, ReasonMaintenanceSuccess and empty message.
func MaintenanceSuccess() metav1.Condition {
	return metav1.Condition{
		Type:               TypeInMaintenance,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonMaintenanceSuccess,
		Message:            "Maintenance concluded successfully",
	}
}

// MaintenanceFailed creates an inactive condition with TypeInMaintenance, ReasonMaintenanceFailure and given message.
func MaintenanceFailed(message string) metav1.Condition {
	return metav1.Condition{
		Type:               TypeInMaintenance,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonMaintenanceFailure,
		Message:            message,
	}
}
