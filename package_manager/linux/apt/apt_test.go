package apt

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHelperProcess is used by exec.Command mocking - it's not a real test
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

func TestClient_Name(t *testing.T) {
	s := &Client{}
	assert.Equal(t, "apt", s.Name())
}

func TestClient_Install_NonRoot(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() {
		execCommand = origCmd
		osGetuid = origGetuid
	}()

	osGetuid = func() int { return 1000 }

	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Install([]string{"curl", "wget"})

	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"apt", "install", "-y", "curl", "wget"}, capturedArgs)
}

func TestClient_Install_Root(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() {
		execCommand = origCmd
		osGetuid = origGetuid
	}()

	osGetuid = func() int { return 0 }

	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Install([]string{"curl"})

	assert.NoError(t, err)
	assert.Equal(t, "apt", capturedName)
	assert.Equal(t, []string{"install", "-y", "curl"}, capturedArgs)
}

func TestClient_Install_Error(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() {
		execCommand = origCmd
		osGetuid = origGetuid
	}()

	osGetuid = func() int { return 0 }
	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Install([]string{"nonexistent-pkg"})

	assert.Error(t, err)
}

func TestParseAptSearch(t *testing.T) {
	output := `curl - command line tool for transferring data with URL syntax
libcurl4 - easy-to-use client-side URL transfer library
`
	results := parseAptSearch(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "command line tool for transferring data with URL syntax", results[0].Description)
	assert.Equal(t, "", results[0].Version)
	assert.Equal(t, "libcurl4", results[1].Name)
}

func TestParseAptSearch_Empty(t *testing.T) {
	results := parseAptSearch("")
	assert.Empty(t, results)
}
