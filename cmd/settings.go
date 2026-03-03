package cmd

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// SettingsCmd displays lay configuration and environment variables
var SettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "List the paths and other environment variables for Lay",
	Run: func(cmd *cobra.Command, args []string) {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Setting", "Value"})

		layPM := os.Getenv("LAY_PACKAGE_MANAGER")
		if layPM == "" {
			layPM = "(not set)"
		}
		table.Append([]string{"LAY_PACKAGE_MANAGER", layPM})

		layRT := os.Getenv("LAY_CONTAINER_RUNTIME")
		if layRT == "" {
			layRT = "(not set)"
		}
		table.Append([]string{"LAY_CONTAINER_RUNTIME", layRT})

		layBP := os.Getenv("LAY_BINARY_PATH")
		if layBP == "" {
			layBP = "(not set — default: ~/.local/bin)"
		}
		table.Append([]string{"LAY_BINARY_PATH", layBP})

		table.Render()
	},
}
