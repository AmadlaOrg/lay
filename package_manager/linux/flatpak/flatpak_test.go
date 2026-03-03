package flatpak

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
	assert.Equal(t, "flatpak", s.Name())
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
	err := s.Install([]string{"org.videolan.VLC"})
	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"flatpak", "install", "-y", "org.videolan.VLC"}, capturedArgs)
}

func TestClient_Install_Root(t *testing.T) {
	origCmd := execCommand
	origGetuid := osGetuid
	defer func() { execCommand = origCmd; osGetuid = origGetuid }()

	osGetuid = func() int { return 0 }
	var capturedName string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		return fakeExecCommand(0)(name, args...)
	}

	s := &Client{}
	err := s.Install([]string{"org.videolan.VLC"})
	assert.NoError(t, err)
	assert.Equal(t, "flatpak", capturedName)
}

func TestParseFlatpakSearch(t *testing.T) {
	output := "VLC\tVLC media player\torg.videolan.VLC\t3.0.20\tstable\tflathub\n"
	results := parseFlatpakSearch(output)
	assert.Len(t, results, 1)
	assert.Equal(t, "VLC", results[0].Name)
	assert.Equal(t, "3.0.20", results[0].Version)
	assert.Equal(t, "VLC media player", results[0].Description)
}

func TestParseFlatpakSearch_Empty(t *testing.T) {
	results := parseFlatpakSearch("")
	assert.Empty(t, results)
}
