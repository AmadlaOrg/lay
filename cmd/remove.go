package cmd

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/lay/output"
	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/spf13/cobra"
)

// removeCmd removes packages using the detected system package manager
var removeCmd = &cobra.Command{
	Use:   "remove [packages...]",
	Short: "Remove packages using the system package manager",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		detector := newDetector()
		out := StdoutWriter()
		if err := runRemove(detector, newManagerByName, managerFlag, args, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runRemove(detector pm.Detector, managerFn func(string) (pm.Manager, error), override string, packages []string, out *output.Writer) error {
	name, err := detector.Detect(override)
	if err != nil {
		return err
	}

	manager, err := managerFn(name)
	if err != nil {
		return err
	}

	out.Info("Using package manager: %s", manager.Name())

	// Check which packages are actually installed
	var toRemove []string
	for _, pkg := range packages {
		installed, err := manager.IsInstalled(pkg)
		if err != nil {
			// On error, proceed with remove attempt
			toRemove = append(toRemove, pkg)
			continue
		}
		if !installed {
			out.Info("Not installed: %s", pkg)
		} else {
			toRemove = append(toRemove, pkg)
		}
	}

	if len(toRemove) == 0 {
		out.Info("No packages to remove")
		return nil
	}

	if flagDryRun {
		out.Info("[dry-run] Would remove: %v", toRemove)
		return nil
	}

	return manager.Remove(toRemove)
}
