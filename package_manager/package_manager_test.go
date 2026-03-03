package package_manager

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	s := &DetectorService{}

	tests := []struct {
		name            string
		managerOverride string
		envValue        string
		lookPathResults map[string]error
		expected        string
		expectErr       bool
		errContains     string
	}{
		{
			name:            "Flag override found",
			managerOverride: "apt",
			lookPathResults: map[string]error{"apt": nil},
			expected:        "apt",
		},
		{
			name:            "Flag override not found",
			managerOverride: "nonexistent",
			lookPathResults: map[string]error{"nonexistent": errors.New("not found")},
			expectErr:       true,
			errContains:     "specified package manager not found",
		},
		{
			name:            "Env var found",
			envValue:        "dnf",
			lookPathResults: map[string]error{"dnf": nil},
			expected:        "dnf",
		},
		{
			name:            "Env var set but binary not found",
			envValue:        "nonexistent",
			lookPathResults: map[string]error{"nonexistent": errors.New("not found")},
			expectErr:       true,
			errContains:     "LAY_PACKAGE_MANAGER",
		},
		{
			name: "Auto-detect finds apt",
			lookPathResults: map[string]error{
				"apt": nil,
				"dnf": errors.New("not found"),
				"yum": errors.New("not found"),
			},
			expected: "apt",
		},
		{
			name: "Auto-detect finds dnf when no apt",
			lookPathResults: map[string]error{
				"apt": errors.New("not found"),
				"dnf": nil,
				"yum": errors.New("not found"),
			},
			expected: "dnf",
		},
		{
			name: "Auto-detect finds yum when no apt or dnf",
			lookPathResults: map[string]error{
				"apt": errors.New("not found"),
				"dnf": errors.New("not found"),
				"yum": nil,
			},
			expected: "yum",
		},
		{
			name: "Nothing found",
			lookPathResults: map[string]error{
				"apt": errors.New("not found"),
				"dnf": errors.New("not found"),
				"yum": errors.New("not found"),
			},
			expectErr:   true,
			errContains: "no supported package manager found",
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
				if key == "LAY_PACKAGE_MANAGER" {
					return tt.envValue
				}
				return ""
			}
			defer func() { osGetenv = origGetenv }()

			result, err := s.Detect(tt.managerOverride)

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
