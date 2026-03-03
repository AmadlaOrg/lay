package compile

import (
	"fmt"
	"os"
)

// For mocking
var osStat = os.Stat

// BuildSystem defines the interface for a build system
type BuildSystem interface {
	Build(srcDir string, targetDir string) error
	Name() string
}

// Detector detects the build system used by a source directory
type Detector interface {
	Detect(srcDir string, override string) (string, error)
}

// DetectorService detects build systems by marker files
type DetectorService struct{}

// markerFiles maps build system names to their marker files, in priority order
var markerFiles = []struct {
	name    string
	markers []string
}{
	{"autotools", []string{"configure", "configure.ac"}},
	{"cmake", []string{"CMakeLists.txt"}},
	{"meson", []string{"meson.build"}},
	{"cargo", []string{"Cargo.toml"}},
	{"golang", []string{"go.mod"}},
	{"makefile", []string{"Makefile", "makefile", "GNUmakefile"}},
}

// Detect identifies the build system in srcDir. If override is non-empty, it is returned directly.
func (s *DetectorService) Detect(srcDir string, override string) (string, error) {
	if override != "" {
		return override, nil
	}

	for _, entry := range markerFiles {
		for _, marker := range entry.markers {
			path := srcDir + "/" + marker
			if _, err := osStat(path); err == nil {
				return entry.name, nil
			}
		}
	}

	return "", fmt.Errorf("no supported build system detected in %s", srcDir)
}
