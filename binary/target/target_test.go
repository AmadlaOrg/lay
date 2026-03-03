package target

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_Resolve(t *testing.T) {
	tests := []struct {
		name         string
		flagOverride string
		envValue     string
		homeDir      string
		homeErr      error
		expected     string
		expectErr    bool
	}{
		{
			name:         "flag override takes priority",
			flagOverride: "/opt/bin",
			envValue:     "/env/bin",
			homeDir:      "/home/user",
			expected:     "/opt/bin",
		},
		{
			name:     "env var used when no flag",
			envValue: "/env/bin",
			homeDir:  "/home/user",
			expected: "/env/bin",
		},
		{
			name:     "default to ~/.local/bin",
			homeDir:  "/home/user",
			expected: filepath.Join("/home/user", ".local", "bin"),
		},
		{
			name:      "error when home dir fails",
			homeErr:   errors.New("no home"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origGetenv := osGetenv
			origHomeDir := osUserHomeDir
			defer func() {
				osGetenv = origGetenv
				osUserHomeDir = origHomeDir
			}()

			osGetenv = func(key string) string {
				if key == "LAY_BINARY_PATH" {
					return tt.envValue
				}
				return ""
			}
			osUserHomeDir = func() (string, error) {
				return tt.homeDir, tt.homeErr
			}

			s := &Service{}
			result, err := s.Resolve(tt.flagOverride)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestService_Ensure(t *testing.T) {
	origMkdirAll := osMkdirAll
	defer func() { osMkdirAll = origMkdirAll }()

	t.Run("creates directory", func(t *testing.T) {
		var calledPath string
		var calledPerm os.FileMode
		osMkdirAll = func(path string, perm os.FileMode) error {
			calledPath = path
			calledPerm = perm
			return nil
		}

		s := &Service{}
		err := s.Ensure("/some/path")

		assert.NoError(t, err)
		assert.Equal(t, "/some/path", calledPath)
		assert.Equal(t, os.FileMode(0755), calledPerm)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		osMkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("permission denied")
		}

		s := &Service{}
		err := s.Ensure("/root/bin")

		assert.Error(t, err)
	})
}
