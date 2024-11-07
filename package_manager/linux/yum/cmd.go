package yum

import "github.com/spf13/cobra"

var PackageYumCmd = &cobra.Command{
	Use:   "yum",
	Short: "",
	Run:   func(cmd *cobra.Command, args []string) {},
}
