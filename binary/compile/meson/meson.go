package meson

import (
	"fmt"
	"os"
	"os/exec"
)

// For mocking
var execCommand = exec.Command

// Builder implements the meson build system
type Builder struct{}

// Name returns the build system name
func (s *Builder) Name() string {
	return "meson"
}

// Build runs meson setup, ninja build, and ninja install
func (s *Builder) Build(srcDir string, targetDir string) error {
	// meson setup build --prefix=<targetDir>
	setup := execCommand("meson", "setup", "build", "--prefix="+targetDir)
	setup.Dir = srcDir
	setup.Stdout = os.Stdout
	setup.Stderr = os.Stderr
	if err := setup.Run(); err != nil {
		return fmt.Errorf("meson setup failed: %w", err)
	}

	// ninja -C build
	build := execCommand("ninja", "-C", "build")
	build.Dir = srcDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("ninja build failed: %w", err)
	}

	// ninja -C build install
	install := execCommand("ninja", "-C", "build", "install")
	install.Dir = srcDir
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("ninja install failed: %w", err)
	}

	return nil
}
