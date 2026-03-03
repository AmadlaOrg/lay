package yum

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
	assert.Equal(t, "yum", s.Name())
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
	assert.Equal(t, []string{"yum", "install", "-y", "curl", "wget"}, capturedArgs)
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
	assert.Equal(t, "yum", capturedName)
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

func TestParseYumSearch(t *testing.T) {
	output := `========================= Name Exactly Matched: curl =========================
curl.x86_64 : A utility for getting files from remote servers
`
	results := parseYumSearch(output)
	assert.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "A utility for getting files from remote servers", results[0].Description)
}

func TestParseYumSearch_Empty(t *testing.T) {
	results := parseYumSearch("")
	assert.Empty(t, results)
}
