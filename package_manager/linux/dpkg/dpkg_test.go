package dpkg

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
