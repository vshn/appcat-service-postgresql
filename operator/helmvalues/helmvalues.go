package helmvalues

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"k8s.io/apimachinery/pkg/runtime"
)

// We are not doing `type V map[string]interface{}` as this breaks type assertion when deep-merging.
// For example, if m is `V` then ok in `vals, ok := m.(map[string]interface{}` will be false.

// V contains the Helm values tree.
type V = map[string]interface{}

// Unmarshal  the values from a runtime.RawExtension.
func Unmarshal(from runtime.RawExtension, into *V) error {
	newMap := map[string]interface{}{}
	err := json.Unmarshal(from.Raw, &newMap)
	if err == nil {
		*into = newMap
	}
	return err
}

// MustUnmarshal is like Unmarshal but panics if there's an error.
func MustUnmarshal(from runtime.RawExtension, into *V) {
	err := Unmarshal(from, into)
	if err != nil {
		panic(fmt.Errorf("cannot unmarshal map: %w", err))
	}
}

// Marshal returns a runtime.RawExtension object.
func Marshal(v V) (runtime.RawExtension, error) {
	raw, err := json.Marshal(v)
	return runtime.RawExtension{Raw: raw}, err
}

// MustMarshal is like Marshal but panics if there's an error.
func MustMarshal(v V) runtime.RawExtension {
	raw, err := Marshal(v)
	if err != nil {
		panic(fmt.Errorf("cannot marshal values: %w", err))
	}
	return raw
}

// Merge deep-merges the given values into this object.
func Merge(values V, into *V) {
	dest := *into
	if dest == nil {
		dest = map[string]interface{}{}
	}
	customMerge(values, dest)
	*into = dest
}

// MustHashSum returns the hash sum of the JSON in the values.
// It panics on errors.
func MustHashSum(from runtime.RawExtension) uint32 {
	h := fnv.New32a()
	_, err := h.Write(from.Raw)
	if err != nil {
		panic(err)
	}
	return h.Sum32()
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
