package cmd

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/lay/output"
	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/spf13/cobra"
)

// updateCmd refreshes the package index
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Refresh the package index",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		detector := newDetector()
		out := StdoutWriter()
		if err := runUpdate(detector, newManagerByName, managerFlag, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runUpdate(detector pm.Detector, managerFn func(string) (pm.Manager, error), override string, out *output.Writer) error {
	name, err := detector.Detect(override)
	if err != nil {
		return err
	}

	manager, err := managerFn(name)
	if err != nil {
		return err
	}

	out.Info("Using package manager: %s", manager.Name())

	if flagDryRun {
		out.Info("[dry-run] Would refresh package index using %s", manager.Name())
		return nil
	}

	return manager.Update()
}
