package golang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	assert.Equal(t, "golang", s.Name())
}

func TestBuilder_Build(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	srcDir := t.TempDir()
	targetDir := t.TempDir()
	srcName := filepath.Base(srcDir)

	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return fakeExecCommand(0)(name, args...)
	}

	s := &Builder{}
	err := s.Build(srcDir, targetDir)

	assert.NoError(t, err)
	assert.Equal(t, "go", capturedName)
	assert.Equal(t, []string{"build", "-o", filepath.Join(targetDir, srcName), "."}, capturedArgs)
}

func TestBuilder_Build_Fails(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	srcDir := t.TempDir()

	execCommand = fakeExecCommand(1)

	s := &Builder{}
	err := s.Build(srcDir, "/target")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "go build failed")
}
