package docker

import "github.com/spf13/cobra"

var PackageDockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
