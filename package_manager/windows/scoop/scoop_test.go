package scoop

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
	assert.Equal(t, "scoop", s.Name())
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
	err := s.Install([]string{"curl", "wget"})

	assert.NoError(t, err)
	assert.Equal(t, "scoop", capturedName)
	assert.Equal(t, []string{"install", "curl", "wget"}, capturedArgs)
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

	output := "Name    Version Source\n----    ------- ------\ncurl    8.4.0   main\nwget    1.21    main\n"
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

func TestParseScoopSearch(t *testing.T) {
	output := "Name    Version Source\n----    ------- ------\ncurl    8.4.0   main\nwget    1.21    main\n"
	results := parseScoopSearch(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.4.0", results[0].Version)
}

func TestParseScoopSearch_Empty(t *testing.T) {
	results := parseScoopSearch("")
	assert.Empty(t, results)
}

func TestParseScoopSearch_NoHeader(t *testing.T) {
	// First non-empty line isn't a header — treat as data immediately
	output := "curl    8.4.0   main\nwget    1.21    main\n"
	results := parseScoopSearch(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
}

func TestParseScoopSearch_ExtraSeparators(t *testing.T) {
	// Separator lines after header should be skipped
	output := "Name    Version Source\n----    ------- ------\ncurl    8.4.0   main\n----\nwget    1.21    main\n"
	results := parseScoopSearch(output)
	assert.Len(t, results, 2)
}

func TestParseScoopSearch_SingleFieldLine(t *testing.T) {
	// Lines with fewer than 2 fields should be skipped
	output := "Name    Version Source\n----    ------- ------\ncurl    8.4.0   main\nbadline\n"
	results := parseScoopSearch(output)
	assert.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Name)
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
	assert.Equal(t, "scoop", capturedName)
	assert.Equal(t, []string{"update"}, capturedArgs)
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
	assert.Equal(t, "scoop", capturedName)
	assert.Equal(t, []string{"update", "*"}, capturedArgs)
}

func TestClient_Upgrade_Specific(t *testing.T) {
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
	err := s.Upgrade([]string{"curl", "wget"})

	assert.NoError(t, err)
	assert.Equal(t, "scoop", capturedName)
	assert.Equal(t, []string{"update", "curl", "wget"}, capturedArgs)
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

	output := "Installed apps:\n\nName    Version Source Updated             Info\n----    ------- ------ -------             ----\ncurl    8.5.0   main   2024-01-15 10:00:00\nwget    1.21.4  main   2024-01-15 10:00:00\n"
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

func TestParseScoopList(t *testing.T) {
	output := "Installed apps:\n\nName    Version Source Updated             Info\n----    ------- ------ -------             ----\ncurl    8.5.0   main   2024-01-15 10:00:00\nwget    1.21.4  main   2024-01-15 10:00:00\n"
	results := parseScoopList(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "1.21.4", results[1].Version)
}

func TestParseScoopList_Empty(t *testing.T) {
	results := parseScoopList("")
	assert.Empty(t, results)
}

func TestClient_IsInstalled_True(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(0)

	s := &Client{}
	installed, err := s.IsInstalled("curl")

	assert.NoError(t, err)
	assert.True(t, installed)
}

func TestClient_IsInstalled_False(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

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
	assert.Equal(t, "scoop", capturedName)
	assert.Equal(t, []string{"uninstall", "curl", "wget"}, capturedArgs)
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
	assert.Equal(t, "scoop", svc.Name())
}
