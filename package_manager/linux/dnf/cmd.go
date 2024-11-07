package dnf

import "github.com/spf13/cobra"

var PackageDnfCmd = &cobra.Command{
	Use:   "dnf",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
