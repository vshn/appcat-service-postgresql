package standalone

import (
	"encoding/json"
	"fmt"

	"github.com/imdario/mergo"
	"k8s.io/apimachinery/pkg/runtime"
)

// HelmValues contains the Helm values tree.
type HelmValues map[string]interface{}

// Unmarshal sets the values from a runtime.RawExtension.
func (v *HelmValues) Unmarshal(raw runtime.RawExtension) error {
	newMap := map[string]interface{}{}
	err := json.Unmarshal(raw.Raw, &newMap)
	if err == nil {
		*v = newMap
	}
	return err
}

// Marshal returns a runtime.RawExtension object.
func (v HelmValues) Marshal() (runtime.RawExtension, error) {
	raw, err := json.Marshal(v)
	return runtime.RawExtension{Raw: raw}, err
}

// MustMarshal is like Marshal but panics if there's an error.
func (v HelmValues) MustMarshal() runtime.RawExtension {
	raw, err := v.Marshal()
	if err != nil {
		panic(fmt.Errorf("cannot marshal values: %w", err))
	}
	return raw
}

// MergeWith merges the given values into this object.
// Non-empty objects are overwritten with empty objects if they are present in values.
func (v *HelmValues) MergeWith(values HelmValues) error {
	return mergo.Merge(v, values, mergo.WithOverride, mergo.WithOverwriteWithEmptyValue)
}
