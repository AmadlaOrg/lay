package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AmadlaOrg/lay/output"
	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd lists installed packages
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		detector := newDetector()
		out := StdoutWriter()
		if err := runList(detector, newManagerByName, managerFlag, os.Stdout, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runList(detector pm.Detector, managerFn func(string) (pm.Manager, error), override string, w io.Writer, out *output.Writer) error {
	name, err := detector.Detect(override)
	if err != nil {
		return err
	}

	manager, err := managerFn(name)
	if err != nil {
		return err
	}

	packages, err := manager.List()
	if err != nil {
		return err
	}

	if len(packages) == 0 {
		out.Info("No installed packages found (using %s)", manager.Name())
		return nil
	}

	out.Result(packages, func(w io.Writer) {
		fmt.Fprintf(w, "Installed packages (%s): %d\n\n", manager.Name(), len(packages))

		table := tablewriter.NewWriter(w)
		table.SetHeader([]string{"Name", "Version"})
		table.SetAutoWrapText(false)
		table.SetBorder(false)
		table.SetColumnSeparator(" ")
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, p := range packages {
			table.Append([]string{p.Name, p.Version})
		}
		table.Render()
	})

	return nil
}
