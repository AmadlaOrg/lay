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

	output := "curl.x86_64 : A utility for getting files from remote servers\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.Search("curl")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Name)
}

func TestClient_Search_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.Search("nonexistent")

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

func TestClient_Update(t *testing.T) {
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
	err := s.Update()

	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"yum", "makecache"}, capturedArgs)
}

func TestClient_Update_Error(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() {
		execCommand = origCmd
		osGetuid = origGetuid
	}()

	osGetuid = func() int { return 0 }
	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Update()

	assert.Error(t, err)
}

func TestClient_Upgrade(t *testing.T) {
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
	err := s.Upgrade(nil)

	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"yum", "update", "-y"}, capturedArgs)
}

func TestClient_Upgrade_Specific(t *testing.T) {
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
	err := s.Upgrade([]string{"curl", "wget"})

	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"yum", "update", "-y", "curl", "wget"}, capturedArgs)
}

func TestClient_Upgrade_Error(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() {
		execCommand = origCmd
		osGetuid = origGetuid
	}()

	osGetuid = func() int { return 0 }
	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Upgrade(nil)

	assert.Error(t, err)
}

func TestClient_List(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "curl.x86_64                          7.76.1-26.el9            @baseos\nbash.x86_64                          5.1.8-6.el9              @baseos\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	packages, err := s.List()

	assert.NoError(t, err)
	assert.Len(t, packages, 2)
	assert.Equal(t, "curl", packages[0].Name)
	assert.Equal(t, "7.76.1-26.el9", packages[0].Version)
	assert.Equal(t, "bash", packages[1].Name)
	assert.Equal(t, "5.1.8-6.el9", packages[1].Version)
}

func TestClient_List_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.List()

	assert.Error(t, err)
}

func TestParseYumList(t *testing.T) {
	output := `curl.x86_64                          7.76.1-26.el9            @baseos
bash.x86_64                          5.1.8-6.el9              @baseos
`
	packages := parseYumList(output)
	assert.Len(t, packages, 2)
	assert.Equal(t, "curl", packages[0].Name)
	assert.Equal(t, "7.76.1-26.el9", packages[0].Version)
	assert.Equal(t, "bash", packages[1].Name)
	assert.Equal(t, "5.1.8-6.el9", packages[1].Version)
}

func TestParseYumList_Empty(t *testing.T) {
	packages := parseYumList("")
	assert.Empty(t, packages)
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

func TestClient_Remove_NonRoot(t *testing.T) {
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
	err := s.Remove([]string{"curl", "wget"})

	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"yum", "remove", "-y", "curl", "wget"}, capturedArgs)
}

func TestClient_Remove_Root(t *testing.T) {
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
	err := s.Remove([]string{"curl"})

	assert.NoError(t, err)
	assert.Equal(t, "yum", capturedName)
	assert.Equal(t, []string{"remove", "-y", "curl"}, capturedArgs)
}

func TestClient_Remove_Error(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() {
		execCommand = origCmd
		osGetuid = origGetuid
	}()

	osGetuid = func() int { return 0 }
	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Remove([]string{"nonexistent-pkg"})

	assert.Error(t, err)
}
