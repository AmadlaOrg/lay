package podman

import "github.com/spf13/cobra"

var PackagePodmanCmd = &cobra.Command{
	Use:   "podman",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
