package dpkg

import "github.com/spf13/cobra"

var PackageDpkgCmd = &cobra.Command{
	Use:   "dpkg",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
