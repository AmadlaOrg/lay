package nix

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

func TestClient_Name(t *testing.T) {
	s := &Client{}
	assert.Equal(t, "nix", s.Name())
}

func TestClient_Install(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Install([]string{"curl", "nixpkgs#wget"})
	assert.NoError(t, err)
	assert.Equal(t, "nix", capturedName)
	// curl should get nixpkgs# prefix, wget already has it
	assert.Equal(t, []string{"profile", "install", "nixpkgs#curl", "nixpkgs#wget"}, capturedArgs)
}

func TestParseNixSearch(t *testing.T) {
	output := `* nixpkgs#curl (8.5.0)
  A command line tool for transferring files with URL syntax

* nixpkgs#wget (1.21.4)
  Tool for retrieving files using HTTP, HTTPS, and FTP
`
	results := parseNixSearch(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0", results[0].Version)
	assert.Equal(t, "A command line tool for transferring files with URL syntax", results[0].Description)
	assert.Equal(t, "wget", results[1].Name)
}

func TestParseNixSearch_Empty(t *testing.T) {
	results := parseNixSearch("")
	assert.Empty(t, results)
}
