package autotools

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// For mocking
var execCommand = exec.Command

// Builder implements the autotools build system
type Builder struct{}

// Name returns the build system name
func (s *Builder) Name() string {
	return "autotools"
}

// Build runs ./configure && make && make install with the given prefix
func (s *Builder) Build(srcDir string, targetDir string) error {
	nproc := fmt.Sprintf("%d", runtime.NumCPU())

	// ./configure --prefix=<targetDir>
	configure := execCommand("./configure", "--prefix="+targetDir)
	configure.Dir = srcDir
	configure.Stdout = os.Stdout
	configure.Stderr = os.Stderr
	if err := configure.Run(); err != nil {
		return fmt.Errorf("configure failed: %w", err)
	}

	// make -j<nproc>
	makeCmd := execCommand("make", "-j"+nproc)
	makeCmd.Dir = srcDir
	makeCmd.Stdout = os.Stdout
	makeCmd.Stderr = os.Stderr
	if err := makeCmd.Run(); err != nil {
		return fmt.Errorf("make failed: %w", err)
	}

	// make install
	install := execCommand("make", "install")
	install.Dir = srcDir
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("make install failed: %w", err)
	}

	return nil
}
