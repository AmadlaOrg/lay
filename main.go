package main

import (
	"fmt"
	"github.com/AmadlaOrg/lay/cmd"
	"github.com/spf13/cobra"
	"log"
)

const appName = "lay"
const appTitleName = "lay"
const version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:     appName,
	Short:   appTitleName + " CLI application",
	Version: version,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of " + appName,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(appName + " version " + version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(cmd.CompileCmd)
	rootCmd.AddCommand(cmd.ContainerCmd)
	rootCmd.AddCommand(cmd.PackageCmd)
	rootCmd.AddCommand(cmd.SettingsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		//os.Exit(1)
	}
}
