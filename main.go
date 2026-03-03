package main

import (
	"github.com/AmadlaOrg/LibraryFramework/cli"
	"github.com/AmadlaOrg/lay/cmd"
	"github.com/spf13/cobra"
)

func main() {
	cli.New(
		"lay",
		"Lay",
		"1.0.0",
		func(rootCmd *cobra.Command) {
			rootCmd.AddCommand(cmd.PackageCmd)
			rootCmd.AddCommand(cmd.ContainerCmd)
			rootCmd.AddCommand(cmd.BinaryCmd)
			rootCmd.AddCommand(cmd.SettingsCmd)
		})
}
