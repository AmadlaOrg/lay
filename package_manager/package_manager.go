package package_manager

import (
	"errors"
	"os/exec"

	"github.com/AmadlaOrg/lay/package_manager/types"
)

// SearchResult is re-exported from the types package for convenience
type SearchResult = types.SearchResult

// Manager defines the interface for package manager operations
type Manager interface {
	Name() string
	Install(packages []string) error
	Remove(packages []string) error
	Search(query string) ([]types.SearchResult, error)
	Update() error
	Upgrade(packages []string) error
	List() ([]types.PackageInfo, error)
	IsInstalled(pkg string) (bool, error)
}

// PackageInfo is re-exported from the types package for convenience
type PackageInfo = types.PackageInfo

// DetectorService handles package manager detection
type DetectorService struct{}

// For mocking
var execLookPath = exec.LookPath

// Detect finds the appropriate package manager to use.
// Priority: managerOverride flag > LAY_PACKAGE_MANAGER env var > auto-detect
func (s *DetectorService) Detect(managerOverride string) (string, error) {
	if managerOverride != "" {
		if _, err := execLookPath(managerOverride); err == nil {
			return managerOverride, nil
		}
		return "", errors.New("specified package manager not found: " + managerOverride)
	}

	if envPM := osGetenv("LAY_PACKAGE_MANAGER"); envPM != "" {
		if _, err := execLookPath(envPM); err == nil {
			return envPM, nil
		}
		return "", errors.New("LAY_PACKAGE_MANAGER set to '" + envPM + "' but binary not found")
	}

	// Auto-detect: try in priority order
	for _, name := range []string{"apt", "dnf", "yum", "pacman", "zypper", "apk", "nix", "snap", "flatpak", "brew", "choco", "scoop", "winget"} {
		if _, err := execLookPath(name); err == nil {
			return name, nil
		}
	}

	return "", errors.New("no supported package manager found")
}
