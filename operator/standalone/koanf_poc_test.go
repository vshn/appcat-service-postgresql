package standalone

import (
	"encoding/json"
	"testing"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKoanfMergeEmptyMap(t *testing.T) {
	t.SkipNow()
	load := func(k *koanf.Koanf, m map[string]interface{}) {
		require.NoError(t, k.Load(confmap.Provider(m, ""), nil), "load maps")
	}
	logJson := func(k *koanf.Koanf) {
		b, err := json.Marshal(k.Raw())
		require.NoError(t, err, "json marshal")
		t.Log(string(b))
	}

	srcKoanf := koanf.New("")
	src := map[string]interface{}{
		"mergeWithEmpty": map[string]interface{}{
			"not empty": "but should become empty when merging (maybe add an option?)",
		},
		"nilMap":   "not nil, but expected to become nil without StrictMerge",
		"emptyMap": map[string]interface{}{},
	}
	load(srcKoanf, src)
	logJson(srcKoanf)
	/* Prints
	{"emptyMap":{},"mergeWithEmpty":{"not empty":"but should become empty when merging (maybe add an option?)"},"nilMap":"not nil, but expected to become nil without StrictMerge"}
	*/

	mergeKoanf := koanf.New("")
	mergeWith := map[string]interface{}{
		"key":            "value",
		"mergeWithEmpty": map[string]interface{}{},
		"nilMap":         nil,
	}

	load(mergeKoanf, mergeWith)
	logJson(mergeKoanf)
	/* Prints
	{"key":"value","mergeWithEmpty":{},"nilMap":null}
	*/

	require.NoError(t, srcKoanf.Merge(mergeKoanf), "merge")
	assert.Equal(t, map[string]interface{}{
		"key":            "value",
		"mergeWithEmpty": map[string]interface{}{},
		"nilMap":         nil,
		"emptyMap":       map[string]interface{}{},
	}, srcKoanf.Raw(), "compare maps")

	/* Prints
	Expected :map[string]interface {}{"emptyMap":map[string]interface {}{}, "key":"value", "mergeWithEmpty":map[string]interface {}{}, "nilMap":interface {}(nil)}
	Actual   :map[string]interface {}{"emptyMap":map[string]interface {}{}, "key":"value", "mergeWithEmpty":map[string]interface {}{"not empty":"but should become empty when merging (maybe add an option?)"}, "nilMap":interface {}(nil)}
	*/

	logJson(srcKoanf)
	/* Prints
	{"emptyMap":{},"key":"value","mergeWithEmpty":{"not empty":"but should become empty when merging (maybe add an option?)"},"nilMap":null}
	*/
}
