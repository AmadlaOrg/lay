package compile

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// For mocking
var execCommand = exec.Command

// CloneOrDownload resolves source to a local directory.
// If source is a URL or GitHub shorthand, it clones into a temp dir.
// If source is a local path, it returns it directly.
// Returns srcDir, cleanup function, and error.
func CloneOrDownload(source string) (string, func(), error) {
	noop := func() {}

	// Check if it's a local path
	if info, err := osStat(source); err == nil && info.IsDir() {
		return source, noop, nil
	}

	// Treat as git URL or GitHub shorthand
	gitURL := source
	if !strings.Contains(source, "://") && !strings.HasSuffix(source, ".git") {
		// Assume GitHub shorthand
		gitURL = "https://github.com/" + source + ".git"
	}

	tmpDir, err := os.MkdirTemp("", "lay-compile-*")
	if err != nil {
		return "", noop, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cleanup := func() { os.RemoveAll(tmpDir) }

	cmd := execCommand("git", "clone", "--depth", "1", gitURL, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		cleanup()
		return "", noop, fmt.Errorf("failed to clone %s: %w", gitURL, err)
	}

	return tmpDir, cleanup, nil
}
