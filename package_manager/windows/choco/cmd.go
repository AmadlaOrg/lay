package choco

import "github.com/spf13/cobra"

var PackageChocoCmd = &cobra.Command{
	Use:   "choco",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
