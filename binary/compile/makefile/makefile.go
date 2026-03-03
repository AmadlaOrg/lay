package makefile

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// For mocking
var execCommand = exec.Command

// Builder implements the makefile build system
type Builder struct{}

// Name returns the build system name
func (s *Builder) Name() string {
	return "makefile"
}

// Build runs make && make install PREFIX=<targetDir>
func (s *Builder) Build(srcDir string, targetDir string) error {
	nproc := fmt.Sprintf("%d", runtime.NumCPU())

	// make -j<nproc>
	makeCmd := execCommand("make", "-j"+nproc)
	makeCmd.Dir = srcDir
	makeCmd.Stdout = os.Stdout
	makeCmd.Stderr = os.Stderr
	if err := makeCmd.Run(); err != nil {
		return fmt.Errorf("make failed: %w", err)
	}

	// make install PREFIX=<targetDir>
	install := execCommand("make", "install", "PREFIX="+targetDir)
	install.Dir = srcDir
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("make install failed: %w", err)
	}

	return nil
}
