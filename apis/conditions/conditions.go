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
	ReasonNotReady               = "NotAvailable"
	ReasonProgressing            = "ProgressingResource"
)

const (
	// TypeInMaintenance identifies a condition related to maintenance.
	TypeInMaintenance = "InMaintenance"
	// TypeReady indicates that an instance is ready to serve.
	TypeReady = "Ready"
	// TypeProgressing indicates that an instance is being updated.
	TypeProgressing = "Progressing"
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

// NotReady creates a condition with TypeReady, ReasonReady and empty message.
func NotReady() metav1.Condition {
	return metav1.Condition{
		Type:               TypeReady,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonNotReady,
	}
}

// Progressing creates a condition with TypeProgressing, ReasonReady and empty message.
func Progressing() metav1.Condition {
	return metav1.Condition{
		Type:               TypeProgressing,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonProgressing,
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
