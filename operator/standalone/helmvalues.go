package standalone

import (
	"encoding/json"
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
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
func (v *HelmValues) MergeWith(values HelmValues) error {
	dstK, err := loadKoanf(*v)
	if err != nil {
		return err
	}
	srcK, err := loadKoanf(values)
	if err != nil {
		return err
	}
	err = dstK.Merge(srcK)
	if err != nil {
		return err
	}
	*v = dstK.Raw()
	return nil
}

func loadKoanf(v map[string]interface{}) (*koanf.Koanf, error) {
	k := koanf.New(".")
	err := k.Load(confmap.Provider(v, ""), nil)
	return k, err
}