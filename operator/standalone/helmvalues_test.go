package standalone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHelmValues_Marshal(t *testing.T) {
	tests := map[string]struct {
		givenMap      HelmValues
		expectedJSON  string
		expectedError string
	}{
		"GivenNilMap_ThenReturnNull": {
			givenMap:     nil,
			expectedJSON: "null",
		},
		"GivenEmptyMap_ThenReturnEmptyObject": {
			givenMap:     HelmValues{},
			expectedJSON: "{}",
		},
		"GivenMapWithValues_ThenReturnJSONObject": {
			givenMap:     HelmValues{"key": "value"},
			expectedJSON: `{"key":"value"}`,
		},
		"GivenMapWithNestedMap_ThenReturnJSONObject": {
			givenMap:     HelmValues{"key": HelmValues{"nested": "value"}},
			expectedJSON: `{"key":{"nested":"value"}}`,
		},
		"GivenMapWithEmptyNestedMap_ThenReturnEmptyJSONObject": {
			givenMap:     HelmValues{"key": HelmValues{}},
			expectedJSON: `{"key":{}}`,
		},
		"GivenMapWithNilValue_ThenReturnNull": {
			givenMap:     HelmValues{"key": nil},
			expectedJSON: `{"key":null}`,
		},
		"GivenMapWithSlice_ThenReturnArray": {
			givenMap:     HelmValues{"array": []string{"string"}},
			expectedJSON: `{"array":["string"]}`,
		},
		"GivenMapWithEmptySlice_ThenReturnEmptyArray": {
			givenMap:     HelmValues{"array": []interface{}{}},
			expectedJSON: `{"array":[]}`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := tc.givenMap.Marshal()
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
		givenMap      HelmValues
		givenRawJSON  string
		expectedMap   HelmValues
		expectedError string
	}{
		"GivenNull_ThenKeepEmpty": {
			givenMap:     nil,
			givenRawJSON: "null",
			expectedMap:  nil,
		},
		"GivenEmptyMap_WhenUnmarshallingEmptyObject_ThenReturnEmptyObject": {
			givenMap:     HelmValues{},
			givenRawJSON: "{}",
			expectedMap:  HelmValues{},
		},
		"GivenNilMap_WhenUnmarshallingEmptyObject_ThenReturnEmptyMap": {
			givenMap:     nil,
			givenRawJSON: "{}",
			expectedMap:  HelmValues{},
		},
		"GivenMapWithValues_WhenUnmarshallingEmptyObject_ThenDeleteExistingValues": {
			givenMap: HelmValues{
				"key": "value",
			},
			givenRawJSON: `{}`,
			expectedMap:  HelmValues{},
		},
		"GivenEmptyMap_WhenUnmarshallingObject_ThenCreateValues": {
			givenMap:     HelmValues{},
			givenRawJSON: `{"key":"value"}`,
			expectedMap:  HelmValues{"key": "value"},
		},
		"GivenEmptyMap_WhenUnmarshallingNestedObject_ThenCreateNestedObject": {
			givenMap:     HelmValues{},
			givenRawJSON: `{"key":{"nested":"value"}}`,
			expectedMap:  HelmValues{"key": map[string]interface{}{"nested": "value"}},
		},
		"GivenEmptyMap_WhenUnmarshallingNestedEmptyObject_ThenSetToEmpty": {
			givenMap:     HelmValues{},
			givenRawJSON: `{"key":{}}`,
			expectedMap:  HelmValues{"key": map[string]interface{}{}},
		},
		"GivenEmptyMap_WhenUnmarshallingNestedNullObject_ThenSetToNil": {
			givenMap:     HelmValues{},
			givenRawJSON: `{"key":null}`,
			expectedMap:  HelmValues{"key": nil},
		},
		"GivenEmptyMap_WhenUnmarshallingEmptyArrays_ThenSetToEmptyArray": {
			givenMap:     HelmValues{},
			givenRawJSON: `{"array":[]}`,
			expectedMap:  HelmValues{"array": []interface{}{}},
		},
		"GivenEmptyMap_WhenUnmarshallingArrays_ThenSetToArray": {
			givenMap:     HelmValues{},
			givenRawJSON: `{"array":["string"]}`,
			expectedMap:  HelmValues{"array": []interface{}{"string"}},
		},
		"GivenNonEmptyMap_WhenUnmarshalling_ThenSetNewValues": {
			givenMap:     HelmValues{"key": "existing"},
			givenRawJSON: `{"key":"value"}`,
			expectedMap:  HelmValues{"key": "value"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.givenMap.Unmarshal(runtime.RawExtension{Raw: []byte(tc.givenRawJSON)})
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
		givenMap      HelmValues
		mergeWith     HelmValues
		expectedMap   HelmValues
		expectedError string
	}{
		"GivenNilMap_ThenCreateFromArgument": {
			givenMap:    nil,
			mergeWith:   HelmValues{"key": "value"},
			expectedMap: HelmValues{"key": "value"},
		},
		"GivenEmptyMap_ThenCreateFromArgument": {
			givenMap:    HelmValues{},
			mergeWith:   HelmValues{"key": "value"},
			expectedMap: HelmValues{"key": "value"},
		},
		"GivenMapWithExistingValue_WhenNilObjectMerged_ThenOverwriteExistingWithNilObject": {
			givenMap:    HelmValues{"key": map[string]interface{}{"nested": "value"}},
			mergeWith:   HelmValues{"key": nil},
			expectedMap: HelmValues{"key": nil},
		},
		/*
			TODO: This currently fails since key isn't overwritten with an empty object.
			This might later become a bug when trying to overwrite an existing object with an explicitly empty object.

			"GivenMapWithExistingValue_WhenEmptyObjectMerged_ThenOverwriteExistingWithEmptyObject": {
				givenMap:    HelmValues{"key": map[string]interface{}{"nested": "value"}},
				mergeWith:   HelmValues{"key": map[string]interface{}{}},
				expectedMap: HelmValues{"key": map[string]interface{}{}},
			},
		*/
		"GivenMapWithExistingValue_WhenObjectMerged_ThenKeepExistingKeys": {
			givenMap:    HelmValues{"key": map[string]interface{}{"nested": "value"}},
			mergeWith:   HelmValues{"key": map[string]interface{}{"another": "value2"}},
			expectedMap: HelmValues{"key": map[string]interface{}{"nested": "value", "another": "value2"}},
		},
		"GivenMapWithExistingValue_WhenObjectHasNestedKeys_ThenOverwriteExistingKeys": {
			givenMap:    HelmValues{"key": "value"},
			mergeWith:   HelmValues{"key": map[string]interface{}{"another": "value2"}},
			expectedMap: HelmValues{"key": map[string]interface{}{"another": "value2"}},
		},
		"GivenMapWithExistingArray_WhenMergingArray_ThenOverwriteExistingWithNewArray": {
			givenMap:    HelmValues{"array": []string{"string"}},
			mergeWith:   HelmValues{"array": []string{"overwrite"}},
			expectedMap: HelmValues{"array": []string{"overwrite"}},
		},
		"GivenMapWithExistingArray_WhenMergingEmptyArray_ThenOverwriteExistingWithEmptyArray": {
			givenMap:    HelmValues{"array": []string{"string"}},
			mergeWith:   HelmValues{"array": []string{}},
			expectedMap: HelmValues{"array": []string{}},
		},
		"GivenMap_WhenMergingArray_ThenCreateNewArray": {
			givenMap:    HelmValues{},
			mergeWith:   HelmValues{"array": []string{"value"}},
			expectedMap: HelmValues{"array": []string{"value"}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.givenMap.MergeWith(tc.mergeWith)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError, "merge error")
				return
			}
			require.NoError(t, err, "merge error")
			assert.Equal(t, tc.expectedMap, tc.givenMap)
		})
	}
}
