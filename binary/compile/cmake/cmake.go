package cmake

import (
	"fmt"
	"os"
	"os/exec"
)

// For mocking
var execCommand = exec.Command

// Builder implements the cmake build system
type Builder struct{}

// Name returns the build system name
func (s *Builder) Name() string {
	return "cmake"
}

// Build runs cmake configure, build, and install
func (s *Builder) Build(srcDir string, targetDir string) error {
	// cmake -B build -DCMAKE_INSTALL_PREFIX=<targetDir>
	configure := execCommand("cmake", "-B", "build", "-DCMAKE_INSTALL_PREFIX="+targetDir)
	configure.Dir = srcDir
	configure.Stdout = os.Stdout
	configure.Stderr = os.Stderr
	if err := configure.Run(); err != nil {
		return fmt.Errorf("cmake configure failed: %w", err)
	}

	// cmake --build build
	build := execCommand("cmake", "--build", "build")
	build.Dir = srcDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("cmake build failed: %w", err)
	}

	// cmake --install build
	install := execCommand("cmake", "--install", "build")
	install.Dir = srcDir
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("cmake install failed: %w", err)
	}

	return nil
}
