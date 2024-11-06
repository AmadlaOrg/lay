package scoop

import "github.com/spf13/cobra"

var PackageScoopCmd = &cobra.Command{
	Use:   "scoop",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
