package cmd

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/lay/output"
	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/spf13/cobra"
)

// upgradeCmd upgrades installed packages
var upgradeCmd = &cobra.Command{
	Use:   "upgrade [packages...]",
	Short: "Upgrade installed packages (all if none specified)",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		detector := newDetector()
		out := StdoutWriter()
		if err := runUpgrade(detector, newManagerByName, managerFlag, args, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runUpgrade(detector pm.Detector, managerFn func(string) (pm.Manager, error), override string, packages []string, out *output.Writer) error {
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
		if len(packages) == 0 {
			out.Info("[dry-run] Would upgrade all packages using %s", manager.Name())
		} else {
			out.Info("[dry-run] Would upgrade packages: %v", packages)
		}
		return nil
	}

	return manager.Upgrade(packages)
}
