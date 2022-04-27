package v1alpha1

// MonitoringEnabledInstance is a reusable type meant for API spec composition.
type MonitoringEnabledInstance struct {
	Monitoring Monitoring `json:"monitoring"`
}

// SlaType identifies the type of SLA options.
type SlaType string

const (
	// SlaBestEffort deploys the metrics stack.
	SlaBestEffort SlaType = "BestEffort"
	// SlaGuaranteed deploys the metrics stack and Prometheus alert rules.
	SlaGuaranteed SlaType = "Guaranteed"
)

// Monitoring contains the settings that control various aspects of metrics and alerting.
type Monitoring struct {
	//+kubebuilder:default=BestEffort
	//+kubebuilder:validation:Enum=BestEffort;Guaranteed

	// SLA contains the SLA name under which the instance runs on.
	SLA SlaType `json:"sla"`
}

// String implements fmt.Stringer.
func (t SlaType) String() string {
	return string(t)
}
