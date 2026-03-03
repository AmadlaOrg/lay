package container

import (
	"fmt"

	"github.com/AmadlaOrg/lay/container/docker"
	"github.com/AmadlaOrg/lay/container/podman"
)

// NewRuntimeByName returns a Runtime implementation for the given name
func NewRuntimeByName(name string) (Runtime, error) {
	switch name {
	case "docker":
		return docker.NewService(), nil
	case "podman":
		return podman.NewService(), nil
	default:
		return nil, fmt.Errorf("unsupported container runtime: %s", name)
	}
}
