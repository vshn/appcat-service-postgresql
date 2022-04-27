package v1alpha1

import (
	"fmt"
)

// DeferrableMaintenance is a reusable type meant for API spec composition.
type DeferrableMaintenance struct {
	UpdatePolicy UpdatePolicy `json:"updatePolicy"`
}

type VersionSelector struct {
	// Major is the major version part.
	Major string `json:"major"`
	// Minor is the minor version part.
	// Some applications consider the minor version part also as part of a major version with breaking changes between minor versions.
	Minor string `json:"minor,omitempty"`
}

type UpdatePolicy struct {
	// Version chooses the version (range) in which the instance is being allowed to do automatic updates.
	Version           VersionSelector   `json:"version"`
	MaintenanceWindow MaintenanceWindow `json:"maintenanceWindow"`
}

type MaintenanceWindow struct {
	Start MaintenanceWindowSelector `json:"start"`
}

type Weekday string

type MaintenanceWindowSelector struct {
	//+kubebuilder:validation:Enum=Sunday;Monday;Tuesday;Wednesday;Thursday;Friday;Saturday

	// Weekday is the day-of-week.
	// It accepts a value of this collection: [Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday].
	// If this is empty, a random Weekday is assigned.
	Weekday Weekday `json:"weekday,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=23

	// Hour defines the starting range of the window.
	// The window is 4h long and the exact starting time is randomized within the 4 hours.
	// If left undefined, a random hour is assigned.
	//
	// Note that this doesn't mean the maintenance will be applied within the 4h window.
	Hour *int `json:"hour,omitempty"`
}

// String implements fmt.Stringer.
// It returns the version string in the format of `major.minor` as-is.
func (s VersionSelector) String() string {
	return fmt.Sprintf("%s.%s", s.Major, s.Minor)
}
