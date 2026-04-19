package cmd

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/lay/output"
	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/spf13/cobra"
)

// For testability
var (
	newDetector      = pm.New
	newManagerByName = pm.NewManagerByName
	osExit           = os.Exit
)

var managerFlag string

// installCmd installs packages using the detected system package manager
var installCmd = &cobra.Command{
	Use:   "install [packages...]",
	Short: "Install packages using the system package manager",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		detector := newDetector()
		out := StdoutWriter()
		if err := runInstall(detector, newManagerByName, managerFlag, args, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runInstall(detector pm.Detector, managerFn func(string) (pm.Manager, error), override string, packages []string, out *output.Writer) error {
	name, err := detector.Detect(override)
	if err != nil {
		return err
	}

	manager, err := managerFn(name)
	if err != nil {
		return err
	}

	out.Info("Using package manager: %s", manager.Name())

	// Idempotency: check which packages are already installed
	var toInstall []string
	for _, pkg := range packages {
		installed, err := manager.IsInstalled(pkg)
		if err != nil {
			// On error, proceed with install attempt
			toInstall = append(toInstall, pkg)
			continue
		}
		if installed {
			out.Info("Already installed: %s", pkg)
		} else {
			toInstall = append(toInstall, pkg)
		}
	}

	if len(toInstall) == 0 {
		out.Info("All packages already installed")
		return nil
	}

	if flagDryRun {
		out.Info("[dry-run] Would install: %v", toInstall)
		return nil
	}

	return manager.Install(toInstall)
}
