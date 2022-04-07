package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleCommand_Validate(t *testing.T) {
	tests := map[string]struct {
		givenExampleFlag string
		expectedError    string
	}{
		// TODO: test cases
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// arrange...
			command := operatorCommand{}
			ctx := newAppContext(t)

			// act...
			err := command.validate(ctx)

			// assert...
			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			assert.NoError(t, err)
		})
	}
}
