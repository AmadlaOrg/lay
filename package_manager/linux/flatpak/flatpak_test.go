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

	output := "VLC\tVLC media player\torg.videolan.VLC\t3.0.20\tstable\tflathub\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.Search("vlc")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "VLC", results[0].Name)
	assert.Equal(t, "3.0.20", results[0].Version)
}

func TestClient_Search_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.Search("nonexistent")

	assert.Error(t, err)
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
	assert.Equal(t, []string{"flatpak", "update", "--appstream", "-y"}, capturedArgs)
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
	err := s.Upgrade([]string{})

	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"flatpak", "update", "-y"}, capturedArgs)
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
	err := s.Upgrade([]string{"org.videolan.VLC", "org.mozilla.firefox"})

	assert.NoError(t, err)
	assert.Equal(t, "sudo", capturedName)
	assert.Equal(t, []string{"flatpak", "update", "-y", "org.videolan.VLC", "org.mozilla.firefox"}, capturedArgs)
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
	err := s.Upgrade([]string{})

	assert.Error(t, err)
}

func TestClient_List(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	output := "Firefox\t119.0\nLibreOffice\t7.6.3\n"
	execCommand = fakeExecCommandWithOutput(output)

	s := &Client{}
	results, err := s.List()

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Firefox", results[0].Name)
	assert.Equal(t, "119.0", results[0].Version)
	assert.Equal(t, "LibreOffice", results[1].Name)
	assert.Equal(t, "7.6.3", results[1].Version)
}

func TestClient_List_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	_, err := s.List()

	assert.Error(t, err)
}

func TestParseFlatpakList(t *testing.T) {
	output := "Firefox\t119.0\nLibreOffice\t7.6.3\n"
	results := parseFlatpakList(output)
	assert.Len(t, results, 2)
	assert.Equal(t, "Firefox", results[0].Name)
	assert.Equal(t, "119.0", results[0].Version)
	assert.Equal(t, "LibreOffice", results[1].Name)
	assert.Equal(t, "7.6.3", results[1].Version)
}

func TestParseFlatpakList_Empty(t *testing.T) {
	results := parseFlatpakList("")
	assert.Empty(t, results)
}

func TestClient_IsInstalled_True(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(0)

	s := &Client{}
	installed, err := s.IsInstalled("org.videolan.VLC")

	assert.NoError(t, err)
	assert.True(t, installed)
}

func TestClient_IsInstalled_False(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	installed, err := s.IsInstalled("org.nonexistent.App")

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
	err := s.Remove([]string{"org.videolan.VLC"})
	assert.NoError(t, err)
	assert.Equal(t, "flatpak", capturedName)
	assert.Equal(t, []string{"uninstall", "-y", "org.videolan.VLC"}, capturedArgs)
}

func TestClient_Remove_Error(t *testing.T) {
	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = fakeExecCommand(1)

	s := &Client{}
	err := s.Remove([]string{"org.nonexistent.App"})
	assert.Error(t, err)
}
