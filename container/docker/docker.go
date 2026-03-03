package docker

import (
	"os"
	"os/exec"
)

// For mocking
var execCommand = exec.Command

// Client implements the container runtime interface for docker
type Client struct{}

// Name returns the runtime name
func (s *Client) Name() string {
	return "docker"
}

// Exec runs docker with the given arguments, forwarding stdin/stdout/stderr
func (s *Client) Exec(args []string) error {
	cmd := execCommand("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
