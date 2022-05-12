package standalone

import (
	"encoding/json"
	"fmt"

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

// MustUnmarshal is like Unmarshal but panics if there's an error.
func (v *HelmValues) MustUnmarshal(raw runtime.RawExtension) {
	err := v.Unmarshal(raw)
	if err != nil {
		panic(fmt.Errorf("cannot unmarshal map: %w", err))
	}
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

// MergeWith deep-merges the given values into this object.
func (v *HelmValues) MergeWith(values HelmValues) {
	dest := *v
	if dest == nil {
		dest = map[string]interface{}{}
	}
	customMerge(values, dest)
	*v = dest
}

// Copied from github.com/knadh/koanf/maps/maps.go (v1.4.1)
// Modified so that empty maps in map 'a' overwrite existing maps in 'b'.
// (https://github.com/knadh/koanf/blob/516880fe32716d1b03e95a5ba844b6a3c8fba2a1/maps/maps.go#L107)
// See also https://github.com/knadh/koanf/issues/146
func customMerge(a, b map[string]interface{}) {
	for key, val := range a {
		// Does the key exist in the target map?
		// If no, add it and move on.
		bVal, ok := b[key]
		if !ok {
			b[key] = val
			continue
		}

		if val == nil {
			b[key] = nil
		}

		// If the incoming val is not a map, do a direct merge.
		if _, ok := val.(map[string]interface{}); !ok {
			b[key] = val
			continue
		}

		// The source key and target keys are both maps. Merge them.
		switch v := bVal.(type) {
		case map[string]interface{}:
			// If it is an empty map, set the value to empty.
			if len(val.(map[string]interface{})) == 0 {
				b[key] = map[string]interface{}{}
				continue
			}

			customMerge(val.(map[string]interface{}), v)
		default:
			b[key] = val
		}
	}
}
