package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuntimeByName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectName  string
		expectErr   bool
		errContains string
	}{
		{
			name:       "docker returns docker runtime",
			input:      "docker",
			expectName: "docker",
		},
		{
			name:       "podman returns podman runtime",
			input:      "podman",
			expectName: "podman",
		},
		{
			name:        "unknown name returns error",
			input:       "containerd",
			expectErr:   true,
			errContains: "unsupported container runtime: containerd",
		},
		{
			name:        "empty name returns error",
			input:       "",
			expectErr:   true,
			errContains: "unsupported container runtime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := NewRuntimeByName(tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, rt)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rt)
				assert.Equal(t, tt.expectName, rt.Name())
			}
		})
	}
}
