package cargo

import (
	"fmt"
	"os"
	"os/exec"
)

// For mocking
var execCommand = exec.Command

// Builder implements the cargo (Rust) build system
type Builder struct{}

// Name returns the build system name
func (s *Builder) Name() string {
	return "cargo"
}

// Build runs cargo build --release. The binary is at target/release/<name>.
func (s *Builder) Build(srcDir string, targetDir string) error {
	build := execCommand("cargo", "build", "--release")
	build.Dir = srcDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("cargo build failed: %w", err)
	}

	return nil
}
