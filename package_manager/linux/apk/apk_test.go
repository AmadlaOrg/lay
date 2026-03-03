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
