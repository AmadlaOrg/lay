package rpm

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/AmadlaOrg/lay/package_manager/types"
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
	assert.Equal(t, "rpm", s.Name())
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
	err := s.Install([]string{"./package.rpm"})
	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"rpm", "-i", "./package.rpm"}, capturedArgs)
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
	err := s.Install([]string{"./package.rpm"})
	assert.NoError(t, err)
	assert.Equal(t, "rpm", capturedName)
	assert.Equal(t, []string{"-i", "./package.rpm"}, capturedArgs)
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

	output := "curl\t7.88.1-10.el9\tA utility for getting files from remote servers\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.Search("curl")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "7.88.1-10.el9", results[0].Version)
}

func TestClient_Search_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	// rpm -qa returns exit code 1 when no matches; Search returns nil, nil
	execCommand = fakeExecCommand(1)

	s := &Client{}
	results, err := s.Search("nonexistent")

	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestParseRpmSearch(t *testing.T) {
	output := "curl\t7.88.1-10.el9\tA utility for getting files from remote servers\nbash\t5.2.15-3.el9\tThe GNU Bourne Again shell\n"
	results := parseRpmSearch(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "7.88.1-10.el9", results[0].Version)
	assert.Contains(t, results[0].Description, "getting files from remote servers")
}

func TestParseRpmSearch_Empty(t *testing.T) {
	results := parseRpmSearch("")
	assert.Empty(t, results)
}

func TestClient_Update_NotSupported(t *testing.T) {
	s := &Client{}
	err := s.Update()
	assert.ErrorIs(t, err, types.ErrNotSupported)
}

func TestClient_Upgrade_NotSupported(t *testing.T) {
	s := &Client{}
	err := s.Upgrade([]string{"curl"})
	assert.ErrorIs(t, err, types.ErrNotSupported)
}

func TestClient_List(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "curl\t7.88.1-10.el9\nbash\t5.2.15-3.el9\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.List()

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "7.88.1-10.el9", results[0].Version)
	assert.Equal(t, "bash", results[1].Name)
	assert.Equal(t, "5.2.15-3.el9", results[1].Version)
}

func TestClient_List_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.List()

	assert.Error(t, err)
}

func TestParseRpmList(t *testing.T) {
	output := "curl\t7.88.1-10.el9\nbash\t5.2.15-3.el9\n"
	results := parseRpmList(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "7.88.1-10.el9", results[0].Version)
	assert.Equal(t, "bash", results[1].Name)
	assert.Equal(t, "5.2.15-3.el9", results[1].Version)
}

func TestParseRpmList_Empty(t *testing.T) {
	results := parseRpmList("")
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

func TestClient_Remove_NotSupported(t *testing.T) {
	s := &Client{}
	err := s.Remove([]string{"curl"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rpm does not support package removal")
}
