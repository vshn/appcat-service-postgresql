package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reasons that give more context to conditions.
const (
	ReasonMaintenanceProgressing = "MaintenanceProgressing"
	ReasonMaintenanceSuccess     = "MaintenanceFinishedSuccessfully"
	ReasonMaintenanceFailure     = "MaintenanceFinishedWithError"
	ReasonReady                  = "The resource is ready"
)

const (
	// TypeInMaintenance identifies a condition related to maintenance.
	TypeInMaintenance = "InMaintenance"
	// TypeReady indicates that an instance is ready to serve.
	TypeReady = "Ready"
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
