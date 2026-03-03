package autotools

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("GO_HELPER_PROCESS_FAIL") == "1" {
		fmt.Fprintf(os.Stderr, "simulated failure")
		os.Exit(1)
	}
	os.Exit(0)
}

func fakeExecCommand(exitCode int) func(name string, args ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		if exitCode != 0 {
			cmd.Env = append(cmd.Env, "GO_HELPER_PROCESS_FAIL=1")
		}
		return cmd
	}
}

func TestBuilder_Name(t *testing.T) {
	s := &Builder{}
	assert.Equal(t, "autotools", s.Name())
}

func TestBuilder_Build(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	srcDir := t.TempDir()
	targetDir := t.TempDir()

	var calls []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		calls = append(calls, name)
		return fakeExecCommand(0)(name, args...)
	}

	s := &Builder{}
	err := s.Build(srcDir, targetDir)

	assert.NoError(t, err)
	assert.Equal(t, []string{"./configure", "make", "make"}, calls)
}

func TestBuilder_Build_MakeFails(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	srcDir := t.TempDir()

	callCount := 0
	execCommand = func(name string, args ...string) *exec.Cmd {
		callCount++
		if callCount == 2 {
			return fakeExecCommand(1)(name, args...)
		}
		return fakeExecCommand(0)(name, args...)
	}

	s := &Builder{}
	err := s.Build(srcDir, "/target")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "make failed")
}

func TestBuilder_Build_MakeInstallFails(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	srcDir := t.TempDir()

	callCount := 0
	execCommand = func(name string, args ...string) *exec.Cmd {
		callCount++
		if callCount == 3 {
			return fakeExecCommand(1)(name, args...)
		}
		return fakeExecCommand(0)(name, args...)
	}

	s := &Builder{}
	err := s.Build(srcDir, "/target")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "make install failed")
}

func TestBuilder_Build_ConfigureFails(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	srcDir := t.TempDir()

	execCommand = fakeExecCommand(1)

	s := &Builder{}
	err := s.Build(srcDir, "/target")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configure failed")
}
