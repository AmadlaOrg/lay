package winget

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

func fakeExecCommandWithOutput(output string) func(name string, args ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "GO_HELPER_OUTPUT=" + output}
		return cmd
	}
}

func TestClient_Name(t *testing.T) {
	s := &Client{}
	assert.Equal(t, "winget", s.Name())
}

func TestClient_Install(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	var calls []struct {
		name string
		args []string
	}
	execCommand = func(name string, args ...string) *exec.Cmd {
		calls = append(calls, struct {
			name string
			args []string
		}{name, args})
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Install([]string{"curl", "wget"})

	assert.NoError(t, err)
	assert.Len(t, calls, 2)
	assert.Equal(t, "winget", calls[0].name)
	assert.Equal(t, []string{"install", "curl"}, calls[0].args)
	assert.Equal(t, "winget", calls[1].name)
	assert.Equal(t, []string{"install", "wget"}, calls[1].args)
}

func TestClient_Install_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Install([]string{"nonexistent-pkg"})

	assert.Error(t, err)
}

func TestClient_Search(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "Name   Id            Version  Source\n---------------------------------------\ncurl   curl.curl     8.4.0    winget\nwget   GnuWin.Wget   1.21     winget\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.Search("curl")

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.4.0", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "1.21", results[1].Version)
}

func TestClient_Search_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.Search("nonexistent")

	assert.Error(t, err)
}

func TestParseWingetSearch(t *testing.T) {
	output := "Name   Id            Version  Source\n---------------------------------------\ncurl   curl.curl     8.4.0    winget\nwget   GnuWin.Wget   1.21     winget\n"
	results := parseWingetSearch(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.4.0", results[0].Version)
}

func TestParseWingetSearch_Empty(t *testing.T) {
	results := parseWingetSearch("")
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
	assert.Equal(t, "winget", capturedName)
	assert.Equal(t, []string{"source", "update"}, capturedArgs)
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
	assert.Equal(t, "winget", capturedName)
	assert.Equal(t, []string{"upgrade", "--all"}, capturedArgs)
}

func TestClient_Upgrade_Specific(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	var calls []struct {
		name string
		args []string
	}
	execCommand = func(name string, args ...string) *exec.Cmd {
		calls = append(calls, struct {
			name string
			args []string
		}{name, args})
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Upgrade([]string{"Curl.Curl", "GnuWin32.Wget"})

	assert.NoError(t, err)
	assert.Len(t, calls, 2)
	assert.Equal(t, "winget", calls[0].name)
	assert.Equal(t, []string{"upgrade", "Curl.Curl"}, calls[0].args)
	assert.Equal(t, "winget", calls[1].name)
	assert.Equal(t, []string{"upgrade", "GnuWin32.Wget"}, calls[1].args)
}

func TestClient_Upgrade_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Upgrade([]string{"nonexistent-pkg"})

	assert.Error(t, err)
}

func TestClient_List(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "Name              Id              Version\n-------------------------------------------------\ncurl              Curl.Curl       8.5.0\nwget              GnuWin32.Wget   1.21.4\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.List()

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "1.21.4", results[1].Version)
}

func TestClient_List_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.List()

	assert.Error(t, err)
}

func TestParseWingetList(t *testing.T) {
	output := "Name              Id              Version\n-------------------------------------------------\ncurl              Curl.Curl       8.5.0\nwget              GnuWin32.Wget   1.21.4\n"
	results := parseWingetList(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "1.21.4", results[1].Version)
}

func TestParseWingetList_Empty(t *testing.T) {
	results := parseWingetList("")
	assert.Empty(t, results)
}

func TestClient_IsInstalled_True(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(0)

	s := &Client{}
	installed, err := s.IsInstalled("Curl.Curl")

	assert.NoError(t, err)
	assert.True(t, installed)
}

func TestClient_IsInstalled_False(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	installed, err := s.IsInstalled("nonexistent.pkg")

	assert.NoError(t, err)
	assert.False(t, installed)
}

func TestClient_Remove(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	var calls []struct {
		name string
		args []string
	}
	execCommand = func(name string, args ...string) *exec.Cmd {
		calls = append(calls, struct {
			name string
			args []string
		}{name, args})
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Remove([]string{"curl", "wget"})

	assert.NoError(t, err)
	assert.Len(t, calls, 2)
	assert.Equal(t, "winget", calls[0].name)
	assert.Equal(t, []string{"uninstall", "curl"}, calls[0].args)
	assert.Equal(t, "winget", calls[1].name)
	assert.Equal(t, []string{"uninstall", "wget"}, calls[1].args)
}

func TestClient_Remove_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Remove([]string{"nonexistent-pkg"})

	assert.Error(t, err)
}

func TestNewService(t *testing.T) {
	svc := NewService()
	assert.NotNil(t, svc)
	assert.Equal(t, "winget", svc.Name())
}
