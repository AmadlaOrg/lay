package target

import (
	"os"
	"path/filepath"
)

// For mocking
var (
	osGetenv      = os.Getenv
	osUserHomeDir = os.UserHomeDir
	osMkdirAll    = os.MkdirAll
)

// Target resolves and ensures the binary install directory
type Target interface {
	Resolve(flagOverride string) (string, error)
	Ensure(path string) error
}

// Service implements Target
type Service struct{}

// Resolve returns the target directory for binary installation.
// Priority: flagOverride > LAY_BINARY_PATH env > ~/.local/bin
func (s *Service) Resolve(flagOverride string) (string, error) {
	if flagOverride != "" {
		return flagOverride, nil
	}

	if envPath := osGetenv("LAY_BINARY_PATH"); envPath != "" {
		return envPath, nil
	}

	home, err := osUserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "bin"), nil
}

// Ensure creates the target directory if it does not exist
func (s *Service) Ensure(path string) error {
	return osMkdirAll(path, 0755)
}
