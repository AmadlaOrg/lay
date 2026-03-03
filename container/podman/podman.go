package podman

import (
	"os"
	"os/exec"
)

// For mocking
var execCommand = exec.Command

// Client implements the container runtime interface for podman
type Client struct{}

// Name returns the runtime name
func (s *Client) Name() string {
	return "podman"
}

// Exec runs podman with the given arguments, forwarding stdin/stdout/stderr
func (s *Client) Exec(args []string) error {
	cmd := execCommand("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
