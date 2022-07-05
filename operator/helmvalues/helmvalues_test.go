package helmvalues

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHelmValues_Marshal(t *testing.T) {
	tests := map[string]struct {
		givenMap      V
		expectedJSON  string
		expectedError string
	}{
		"GivenNilMap_ThenReturnNull": {
			givenMap:     nil,
			expectedJSON: "null",
		},
		"GivenEmptyMap_ThenReturnEmptyObject": {
			givenMap:     V{},
			expectedJSON: "{}",
		},
		"GivenMapWithValues_ThenReturnJSONObject": {
			givenMap:     V{"key": "value"},
			expectedJSON: `{"key":"value"}`,
		},
		"GivenMapWithNestedMap_ThenReturnJSONObject": {
			givenMap:     V{"key": V{"nested": "value"}},
			expectedJSON: `{"key":{"nested":"value"}}`,
		},
		"GivenMapWithEmptyNestedMap_ThenReturnEmptyJSONObject": {
			givenMap:     V{"key": V{}},
			expectedJSON: `{"key":{}}`,
		},
		"GivenMapWithNilValue_ThenReturnNull": {
			givenMap:     V{"key": nil},
			expectedJSON: `{"key":null}`,
		},
		"GivenMapWithSlice_ThenReturnArray": {
			givenMap:     V{"array": []string{"string"}},
			expectedJSON: `{"array":["string"]}`,
		},
		"GivenMapWithEmptySlice_ThenReturnEmptyArray": {
			givenMap:     V{"array": []interface{}{}},
			expectedJSON: `{"array":[]}`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := Marshal(tc.givenMap)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, "marshalling error")
				return
			}
			require.NoError(t, err, "marshalling error")
			assert.JSONEq(t, tc.expectedJSON, string(result.Raw))
		})
	}
}

func TestHelmValues_Unmarshal(t *testing.T) {
	tests := map[string]struct {
		givenMap      V
		givenRawJSON  string
		expectedMap   V
		expectedError string
	}{
		"GivenNull_ThenKeepEmpty": {
			givenMap:     nil,
			givenRawJSON: "null",
			expectedMap:  nil,
		},
		"GivenEmptyMap_WhenUnmarshallingEmptyObject_ThenReturnEmptyObject": {
			givenMap:     V{},
			givenRawJSON: "{}",
			expectedMap:  V{},
		},
		"GivenNilMap_WhenUnmarshallingEmptyObject_ThenReturnEmptyMap": {
			givenMap:     nil,
			givenRawJSON: "{}",
			expectedMap:  V{},
		},
		"GivenMapWithValues_WhenUnmarshallingEmptyObject_ThenDeleteExistingValues": {
			givenMap: V{
				"key": "value",
			},
			givenRawJSON: `{}`,
			expectedMap:  V{},
		},
		"GivenEmptyMap_WhenUnmarshallingObject_ThenCreateValues": {
			givenMap:     V{},
			givenRawJSON: `{"key":"value"}`,
			expectedMap:  V{"key": "value"},
		},
		"GivenEmptyMap_WhenUnmarshallingNestedObject_ThenCreateNestedObject": {
			givenMap:     V{},
			givenRawJSON: `{"key":{"nested":"value"}}`,
			expectedMap:  V{"key": map[string]interface{}{"nested": "value"}},
		},
		"GivenEmptyMap_WhenUnmarshallingNestedEmptyObject_ThenSetToEmpty": {
			givenMap:     V{},
			givenRawJSON: `{"key":{}}`,
			expectedMap:  V{"key": map[string]interface{}{}},
		},
		"GivenEmptyMap_WhenUnmarshallingNestedNullObject_ThenSetToNil": {
			givenMap:     V{},
			givenRawJSON: `{"key":null}`,
			expectedMap:  V{"key": nil},
		},
		"GivenEmptyMap_WhenUnmarshallingEmptyArrays_ThenSetToEmptyArray": {
			givenMap:     V{},
			givenRawJSON: `{"array":[]}`,
			expectedMap:  V{"array": []interface{}{}},
		},
		"GivenEmptyMap_WhenUnmarshallingArrays_ThenSetToArray": {
			givenMap:     V{},
			givenRawJSON: `{"array":["string"]}`,
			expectedMap:  V{"array": []interface{}{"string"}},
		},
		"GivenNonEmptyMap_WhenUnmarshalling_ThenSetNewValues": {
			givenMap:     V{"key": "existing"},
			givenRawJSON: `{"key":"value"}`,
			expectedMap:  V{"key": "value"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := Unmarshal(runtime.RawExtension{Raw: []byte(tc.givenRawJSON)}, &tc.givenMap)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, "marshalling error")
				return
			}
			require.NoError(t, err, "marshalling error")
			assert.Equal(t, tc.expectedMap, tc.givenMap)
		})
	}
}

func TestHelmValues_MergeWith(t *testing.T) {
	tests := map[string]struct {
		givenMap      V
		mergeWith     V
		expectedMap   V
		expectedError string
	}{
		"GivenNilMap_ThenCreateFromArgument": {
			givenMap:    nil,
			mergeWith:   V{"key": "value"},
			expectedMap: V{"key": "value"},
		},
		"GivenEmptyMap_ThenCreateFromArgument": {
			givenMap:    V{},
			mergeWith:   V{"key": "value"},
			expectedMap: V{"key": "value"},
		},
		"GivenMapWithExistingValue_WhenNilObjectMerged_ThenOverwriteExistingWithNilObject": {
			givenMap:    V{"key": map[string]interface{}{"nested": "value"}},
			mergeWith:   V{"key": nil},
			expectedMap: V{"key": nil},
		},
		"GivenMapWithExistingValue_WhenEmptyObjectMerged_ThenOverwriteExistingWithEmptyObject": {
			givenMap:    V{"key": map[string]interface{}{"nested": "value"}},
			mergeWith:   V{"key": map[string]interface{}{}},
			expectedMap: V{"key": map[string]interface{}{}},
		},
		"GivenMapWithExistingValue_WhenObjectMerged_ThenKeepExistingKeys": {
			givenMap:    V{"key": map[string]interface{}{"nested": "value"}},
			mergeWith:   V{"key": map[string]interface{}{"another": "value2"}},
			expectedMap: V{"key": map[string]interface{}{"nested": "value", "another": "value2"}},
		},
		"GivenMapWithExistingValue_WhenObjectHasNestedKeys_ThenOverwriteExistingKeys": {
			givenMap:    V{"key": "value"},
			mergeWith:   V{"key": map[string]interface{}{"another": "value2"}},
			expectedMap: V{"key": map[string]interface{}{"another": "value2"}},
		},
		"GivenMapWithExistingArray_WhenMergingArray_ThenOverwriteExistingWithNewArray": {
			givenMap:    V{"array": []string{"string"}},
			mergeWith:   V{"array": []string{"overwrite"}},
			expectedMap: V{"array": []string{"overwrite"}},
		},
		"GivenMapWithExistingArray_WhenMergingEmptyArray_ThenOverwriteExistingWithEmptyArray": {
			givenMap:    V{"array": []string{"string"}},
			mergeWith:   V{"array": []string{}},
			expectedMap: V{"array": []string{}},
		},
		"GivenMap_WhenMergingArray_ThenCreateNewArray": {
			givenMap:    V{},
			mergeWith:   V{"array": []string{"value"}},
			expectedMap: V{"array": []string{"value"}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			Merge(tc.mergeWith, &tc.givenMap)
			assert.Equal(t, tc.expectedMap, tc.givenMap)
			b, err := json.Marshal(tc.givenMap)
			require.NoError(t, err, "json marshal")
			t.Log(string(b))
		})
	}
}

func TestMustMarshal(t *testing.T) {
	assert.NotPanics(t, func() {
		MustMarshal(nil)
	}, "should not panic")
}

func TestMustUnmarshal(t *testing.T) {
	assert.Panics(t, func() {
		MustUnmarshal(runtime.RawExtension{}, nil)
	}, "should panic if JSON is not there")

	assert.Panics(t, func() {
		MustUnmarshal(runtime.RawExtension{Raw: []byte("{}")}, nil)
	}, "should panic if dst is nil")

	assert.NotPanics(t, func() {
		MustUnmarshal(runtime.RawExtension{Raw: []byte("{}")}, &map[string]interface{}{})
	}, "should not panic")
}

func TestMustHashSum(t *testing.T) {
	result := MustHashSum(runtime.RawExtension{Raw: []byte((`{"key":"value"}`))})
	assert.Equal(t, uint32(0x5b495ab5), result)
}
