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
	if output := os.Getenv("GO_HELPER_OUTPUT"); output != "" {
		fmt.Fprint(os.Stdout, output)
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

func fakeExecCommandWithOutput(output string) func(name string, args ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "GO_HELPER_OUTPUT=" + output}
		return cmd
	}
}

func TestClient_Search(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "* nixpkgs#curl (8.5.0)\n  A command line tool for transferring files with URL syntax\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.Search("curl")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0", results[0].Version)
}

func TestClient_Search_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.Search("nonexistent")

	assert.Error(t, err)
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

func TestClient_Update(t *testing.T) {
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
	err := s.Update()

	assert.NoError(t, err)
	assert.Equal(t, "nix", capturedName)
	assert.Equal(t, []string{"flake", "update"}, capturedArgs)
}

func TestClient_Update_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Update()

	assert.Error(t, err)
}

func TestClient_Upgrade(t *testing.T) {
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
	err := s.Upgrade([]string{})

	assert.NoError(t, err)
	assert.Equal(t, "nix", capturedName)
	assert.Equal(t, []string{"profile", "upgrade", ".*"}, capturedArgs)
}

func TestClient_Upgrade_Specific(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	var capturedCalls []struct {
		name string
		args []string
	}
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedCalls = append(capturedCalls, struct {
			name string
			args []string
		}{name, args})
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Upgrade([]string{"curl", "wget"})

	assert.NoError(t, err)
	assert.Len(t, capturedCalls, 2)
	assert.Equal(t, "nix", capturedCalls[0].name)
	assert.Equal(t, []string{"profile", "upgrade", "curl"}, capturedCalls[0].args)
	assert.Equal(t, "nix", capturedCalls[1].name)
	assert.Equal(t, []string{"profile", "upgrade", "wget"}, capturedCalls[1].args)
}

func TestClient_Upgrade_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Upgrade([]string{})

	assert.Error(t, err)
}

func TestClient_List(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "0 flake:nixpkgs#curl github:NixOS/nixpkgs/abc123#curl /nix/store/xxx-curl-8.5.0\n1 flake:nixpkgs#wget github:NixOS/nixpkgs/abc123#wget /nix/store/xxx-wget-1.21\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.List()

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "", results[1].Version)
}

func TestClient_List_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.List()

	assert.Error(t, err)
}

func TestParseNixList(t *testing.T) {
	output := "0 flake:nixpkgs#curl github:NixOS/nixpkgs/abc123#curl /nix/store/xxx-curl-8.5.0\n1 flake:nixpkgs#wget github:NixOS/nixpkgs/abc123#wget /nix/store/xxx-wget-1.21\n"
	results := parseNixList(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "", results[1].Version)
}

func TestParseNixList_Empty(t *testing.T) {
	results := parseNixList("")
	assert.Empty(t, results)
}

func TestClient_IsInstalled_True(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "0 flake:nixpkgs#curl github:NixOS/nixpkgs/abc123#curl /nix/store/xxx-curl-8.5.0\n1 flake:nixpkgs#wget github:NixOS/nixpkgs/abc123#wget /nix/store/xxx-wget-1.21\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	installed, err := s.IsInstalled("curl")

	assert.NoError(t, err)
	assert.True(t, installed)
}

func TestClient_IsInstalled_False(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "0 flake:nixpkgs#curl github:NixOS/nixpkgs/abc123#curl /nix/store/xxx-curl-8.5.0\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	installed, err := s.IsInstalled("nonexistent-pkg")

	assert.NoError(t, err)
	assert.False(t, installed)
}

func TestClient_Remove(t *testing.T) {
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
	err := s.Remove([]string{"curl", "wget"})
	assert.NoError(t, err)
	assert.Equal(t, "nix-env", capturedName)
	assert.Equal(t, []string{"-e", "curl", "wget"}, capturedArgs)
}

func TestClient_Remove_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Remove([]string{"nonexistent-pkg"})
	assert.Error(t, err)
}
