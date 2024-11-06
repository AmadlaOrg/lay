package cmd

import (
	"github.com/AmadlaOrg/lay/container/docker"
	"github.com/AmadlaOrg/lay/container/podman"
	"github.com/spf13/cobra"
)

var ContainerCmd = &cobra.Command{
	Use:   "container",
	Short: "Add software via container",
	//Long:  `A longer description that spans multiple lines and likely`,
}

func init() {
	ContainerCmd.AddCommand(docker.PackageDockerCmd)
	ContainerCmd.AddCommand(podman.PackagePodmanCmd)
}
