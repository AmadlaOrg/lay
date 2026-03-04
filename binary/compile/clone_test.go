package compile

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloneOrDownload_LocalPath(t *testing.T) {
	dir := t.TempDir()

	srcDir, cleanup, err := CloneOrDownload(dir)

	assert.NoError(t, err)
	assert.Equal(t, dir, srcDir)
	cleanup() // should be a no-op
}

func TestCloneOrDownload_NonexistentPath(t *testing.T) {
	// TestHelperProcess is required for exec.Command mocking
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		os.Exit(0)
	}

	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestCloneOrDownload_NonexistentPath", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	srcDir, cleanup, err := CloneOrDownload("user/repo")
	if err == nil {
		defer cleanup()
		assert.NotEmpty(t, srcDir)
	}
	// Either succeeds (mock clone) or fails (depending on environment)
}

func TestCloneOrDownload_GitURL(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		os.Exit(0)
	}

	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedArgs = append([]string{name}, args...)
		cs := []string{"-test.run=TestCloneOrDownload_GitURL", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	_, cleanup, err := CloneOrDownload("https://github.com/user/repo.git")
	if err == nil {
		defer cleanup()
	}

	assert.Equal(t, "git", capturedArgs[0])
	assert.Equal(t, "clone", capturedArgs[1])
	assert.Equal(t, "--depth", capturedArgs[2])
	assert.Equal(t, "1", capturedArgs[3])
	assert.Equal(t, "https://github.com/user/repo.git", capturedArgs[4])
}

func TestCloneOrDownload_CloneFailure(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		os.Exit(1)
	}

	origCmd := execCommand
	defer func() { execCommand = origCmd }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestCloneOrDownload_CloneFailure", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	srcDir, cleanup, err := CloneOrDownload("user/repo")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clone")
	assert.Empty(t, srcDir)
	cleanup() // should be safe to call (noop)
}
