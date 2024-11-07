package apt

import "github.com/spf13/cobra"

var PackageAptCmd = &cobra.Command{
	Use:   "apt",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
