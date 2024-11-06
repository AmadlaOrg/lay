package rpm

import "github.com/spf13/cobra"

var PackageRpmCmd = &cobra.Command{
	Use:   "rpm",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
