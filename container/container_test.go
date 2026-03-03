package container

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	s := &DetectorService{}

	tests := []struct {
		name            string
		runtimeOverride string
		envValue        string
		lookPathResults map[string]error
		expected        string
		expectErr       bool
		errContains     string
	}{
		{
			name:            "Flag override found",
			runtimeOverride: "docker",
			lookPathResults: map[string]error{"docker": nil},
			expected:        "docker",
		},
		{
			name:            "Flag override not found",
			runtimeOverride: "nonexistent",
			lookPathResults: map[string]error{"nonexistent": errors.New("not found")},
			expectErr:       true,
			errContains:     "specified container runtime not found",
		},
		{
			name:            "Env var found",
			envValue:        "docker",
			lookPathResults: map[string]error{"docker": nil},
			expected:        "docker",
		},
		{
			name:            "Env var set but binary not found",
			envValue:        "nonexistent",
			lookPathResults: map[string]error{"nonexistent": errors.New("not found")},
			expectErr:       true,
			errContains:     "LAY_CONTAINER_RUNTIME",
		},
		{
			name: "Auto-detect finds podman first",
			lookPathResults: map[string]error{
				"podman": nil,
				"docker": nil,
			},
			expected: "podman",
		},
		{
			name: "Auto-detect falls back to docker",
			lookPathResults: map[string]error{
				"podman": errors.New("not found"),
				"docker": nil,
			},
			expected: "docker",
		},
		{
			name: "Nothing found",
			lookPathResults: map[string]error{
				"podman": errors.New("not found"),
				"docker": errors.New("not found"),
			},
			expectErr:   true,
			errContains: "no supported container runtime found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock execLookPath
			origLookPath := execLookPath
			execLookPath = func(file string) (string, error) {
				if err, ok := tt.lookPathResults[file]; ok {
					if err != nil {
						return "", err
					}
					return "/usr/bin/" + file, nil
				}
				return "", errors.New("not found")
			}
			defer func() { execLookPath = origLookPath }()

			// Mock osGetenv
			origGetenv := osGetenv
			osGetenv = func(key string) string {
				if key == "LAY_CONTAINER_RUNTIME" {
					return tt.envValue
				}
				return ""
			}
			defer func() { osGetenv = origGetenv }()

			result, err := s.Detect(tt.runtimeOverride)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
