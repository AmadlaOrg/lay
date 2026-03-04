package dpkg

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
	assert.Equal(t, "dpkg", s.Name())
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
	err := s.Install([]string{"./package.deb"})
	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"dpkg", "-i", "./package.deb"}, capturedArgs)
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
	err := s.Install([]string{"./package.deb"})
	assert.NoError(t, err)
	assert.Equal(t, "dpkg", capturedName)
	assert.Equal(t, []string{"-i", "./package.deb"}, capturedArgs)
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

	output := "||/ Name           Version      Architecture Description\n+++-==============-============-============-=================================\nii  curl           7.88.1-10    amd64        command line tool for transferring data\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.Search("curl")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "7.88.1-10", results[0].Version)
}

func TestClient_Search_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	// dpkg -l returns exit code 1 when no matches; Search returns nil, nil
	execCommand = fakeExecCommand(1)

	s := &Client{}
	results, err := s.Search("nonexistent")

	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestParseDpkgSearch(t *testing.T) {
	output := `Desired=Unknown/Install/Remove/Purge/Hold
||/ Name           Version      Architecture Description
+++-==============-============-============-=================================
ii  curl           7.88.1-10    amd64        command line tool for transferring data
ii  libcurl4       7.88.1-10    amd64        easy-to-use client-side URL transfer library
rc  curl-old       7.80.0-1     amd64        old curl version
`
	results := parseDpkgSearch(output)
	assert.Len(t, results, 2) // only "ii" entries
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "7.88.1-10", results[0].Version)
	assert.Contains(t, results[0].Description, "command line tool")
}

func TestParseDpkgSearch_UnicodeHeaders(t *testing.T) {
	output := "┌── Desired=Unknown/Install/Remove/Purge/Hold\n" +
		"│┌─ Status=Not/Inst/Conf-files/Unpacked\n" +
		"││┌ Err?=(none)/Reinst-required\n" +
		"│││ Name                     Version      Architecture Description\n" +
		"├┼┼─════════════════════════─════════════─════════════─══════════════\n" +
		"ii  curl                     8.18.0-2     amd64        command line tool\n"
	results := parseDpkgSearch(output)
	assert.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "8.18.0-2", results[0].Version)
}

func TestParseDpkgSearch_Empty(t *testing.T) {
	results := parseDpkgSearch("")
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

	output := "||/ Name           Version      Architecture Description\n+++-==============-============-============-=================================\nii  curl           7.81.0-1     amd64        command line tool for transferring data\nii  wget           1.21.2-2     amd64        retrieves files from web\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.List()

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "7.81.0-1", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "1.21.2-2", results[1].Version)
}

func TestClient_List_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.List()

	assert.Error(t, err)
}

func TestParseDpkgList(t *testing.T) {
	output := `Desired=Unknown/Install/Remove/Purge/Hold
||/ Name           Version      Architecture Description
+++-==============-============-============-=================================
ii  curl           7.81.0-1     amd64        command line tool for transferring data
ii  wget           1.21.2-2     amd64        retrieves files from web
rc  curl-old       7.80.0-1     amd64        old curl version
`
	results := parseDpkgList(output)
	assert.Len(t, results, 2) // only "ii" entries
	assert.Equal(t, "curl", results[0].Name)
	assert.Equal(t, "7.81.0-1", results[0].Version)
	assert.Equal(t, "wget", results[1].Name)
	assert.Equal(t, "1.21.2-2", results[1].Version)
}

func TestParseDpkgList_Empty(t *testing.T) {
	results := parseDpkgList("")
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
	assert.Contains(t, err.Error(), "dpkg does not support package removal")
}
