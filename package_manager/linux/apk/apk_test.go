package apk

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
	assert.Equal(t, "apk", s.Name())
}

func TestClient_Install_NonRoot(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 1000 }
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
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"apk", "add", "curl"}, capturedArgs)
}

func TestClient_Install_Root(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

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
	assert.Equal(t, "apk", capturedName)
	assert.Equal(t, []string{"add", "curl"}, capturedArgs)
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

	output := "curl-8.5.0-r0 - URL retrieval utility and library\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.Search("curl")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0-r0", results[0].Version)
}

func TestClient_Search_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.Search("nonexistent")

	assert.Error(t, err)
}

func TestParseApkSearch(t *testing.T) {
	output := `curl-8.5.0-r0 - URL retrieval utility and library
wget-1.21.4-r0 - Network utility to retrieve files
`
	results := parseApkSearch(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0-r0", results[0].Version)
	assert.Equal(t, "URL retrieval utility and library", results[0].Description)
}

func TestClient_Update(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 1000 }
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
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"apk", "update"}, capturedArgs)
}

func TestClient_Update_Error(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 0 }
	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Update()
	assert.Error(t, err)
}

func TestClient_Upgrade(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 1000 }
	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Upgrade(nil)
	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"apk", "upgrade"}, capturedArgs)
}

func TestClient_Upgrade_Specific(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 1000 }
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
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"apk", "upgrade", "curl", "wget"}, capturedArgs)
}

func TestClient_Upgrade_Error(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 0 }
	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Upgrade(nil)
	assert.Error(t, err)
}

func TestClient_List(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "curl-8.5.0-r0 x86_64 {curl} (MIT)\nwget-1.21.4-r0 x86_64 {wget} (GPL-3.0-or-later)\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.List()

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0-r0", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "1.21.4-r0", results[1].Version)
}

func TestClient_List_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.List()

	assert.Error(t, err)
}

func TestParseApkList(t *testing.T) {
	output := "curl-8.5.0-r0 x86_64 {curl} (MIT)\nwget-1.21.4-r0 x86_64 {wget} (GPL-3.0-or-later)\n"
	results := parseApkList(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.5.0-r0", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "1.21.4-r0", results[1].Version)
}

func TestParseApkList_Empty(t *testing.T) {
	results := parseApkList("")
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
	installed, err := s.IsInstalled("nonexistent")
	assert.NoError(t, err)
	assert.False(t, installed)
}

func TestSplitApkNameVersion(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion string
	}{
		{"curl-8.5.0-r0", "curl", "8.5.0-r0"},
		{"lib-xml2-2.12.3-r0", "lib-xml2", "2.12.3-r0"},
		{"noversion", "noversion", ""},
	}
	for _, tt := range tests {
		name, version := splitApkNameVersion(tt.input)
		assert.Equal(t, tt.wantName, name, "name for %s", tt.input)
		assert.Equal(t, tt.wantVersion, version, "version for %s", tt.input)
	}
}

func TestClient_Remove_NonRoot(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 1000 }
	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Remove([]string{"curl"})
	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"apk", "del", "curl"}, capturedArgs)
}

func TestClient_Remove_Root(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 0 }
	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Remove([]string{"curl"})
	assert.NoError(t, err)
	assert.Equal(t, "apk", capturedName)
	assert.Equal(t, []string{"del", "curl"}, capturedArgs)
}

func TestClient_Remove_Error(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 0 }
	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Remove([]string{"nonexistent-pkg"})
	assert.Error(t, err)
}
