package cmd

import (
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

// For testability
var osGetenvSettings = os.Getenv

type settingsData struct {
	PackageManager   string `json:"package_manager"`
	ContainerRuntime string `json:"container_runtime"`
	BinaryPath       string `json:"binary_path"`
}

// SettingsCmd displays lay configuration and environment variables
var SettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "List the paths and other environment variables for Lay",
	Run: func(cmd *cobra.Command, args []string) {
		out := StdoutWriter()

		layPM := osGetenvSettings("LAY_PACKAGE_MANAGER")
		layRT := osGetenvSettings("LAY_CONTAINER_RUNTIME")
		layBP := osGetenvSettings("LAY_BINARY_PATH")

		data := settingsData{
			PackageManager:   layPM,
			ContainerRuntime: layRT,
			BinaryPath:       layBP,
		}

		out.Result(data, func(w io.Writer) {
			if data.PackageManager == "" {
				data.PackageManager = "(not set)"
			}
			if data.ContainerRuntime == "" {
				data.ContainerRuntime = "(not set)"
			}
			if data.BinaryPath == "" {
				data.BinaryPath = "(not set — default: ~/.local/bin)"
			}

			table := tablewriter.NewTable(os.Stdout,
				tablewriter.WithHeaderAlignment(tw.AlignLeft),
				tablewriter.WithRowAlignment(tw.AlignLeft),
			)
			table.Header([]string{"Setting", "Value"})
			table.Append([]string{"LAY_PACKAGE_MANAGER", data.PackageManager})
			table.Append([]string{"LAY_CONTAINER_RUNTIME", data.ContainerRuntime})
			table.Append([]string{"LAY_BINARY_PATH", data.BinaryPath})
			table.Render()
		})
	},
}
