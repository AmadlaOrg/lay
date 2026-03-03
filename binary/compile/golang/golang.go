package golang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// For mocking
var execCommand = exec.Command

// Builder implements the Go build system
type Builder struct{}

// Name returns the build system name
func (s *Builder) Name() string {
	return "golang"
}

// Build runs go build -o <targetDir>/<name> .
func (s *Builder) Build(srcDir string, targetDir string) error {
	// Use the source directory name as the binary name
	name := filepath.Base(srcDir)
	outputPath := filepath.Join(targetDir, name)

	build := execCommand("go", "build", "-o", outputPath, ".")
	build.Dir = srcDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	return nil
}
