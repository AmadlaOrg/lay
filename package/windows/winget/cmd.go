package winget

import "github.com/spf13/cobra"

var PackageWingetCmd = &cobra.Command{
	Use:   "winget",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
