package container

import (
	"errors"
	"os/exec"
)

// Runtime defines the interface for container runtime operations
type Runtime interface {
	Exec(args []string) error
	Name() string
}

// DetectorService handles container runtime detection
type DetectorService struct{}

// For mocking
var execLookPath = exec.LookPath

// Detect finds the appropriate container runtime to use.
// Priority: runtimeOverride flag > LAY_CONTAINER_RUNTIME env var > auto-detect (podman first, then docker)
func (s *DetectorService) Detect(runtimeOverride string) (string, error) {
	if runtimeOverride != "" {
		if _, err := execLookPath(runtimeOverride); err == nil {
			return runtimeOverride, nil
		}
		return "", errors.New("specified container runtime not found: " + runtimeOverride)
	}

	if envRT := osGetenv("LAY_CONTAINER_RUNTIME"); envRT != "" {
		if _, err := execLookPath(envRT); err == nil {
			return envRT, nil
		}
		return "", errors.New("LAY_CONTAINER_RUNTIME set to '" + envRT + "' but binary not found")
	}

	// Auto-detect: podman preferred over docker
	for _, name := range []string{"podman", "docker"} {
		if _, err := execLookPath(name); err == nil {
			return name, nil
		}
	}

	return "", errors.New("no supported container runtime found (install docker or podman)")
}
