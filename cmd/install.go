package cmd

import (
	"fmt"
	"os"

	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/spf13/cobra"
)

// For testability
var (
	newDetector      = pm.NewDetectorService
	newManagerByName = pm.NewManagerByName
)

var managerFlag string

// installCmd installs packages using the detected system package manager
var installCmd = &cobra.Command{
	Use:   "install [packages...]",
	Short: "Install packages using the system package manager",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		detector := newDetector()

		name, err := detector.Detect(managerFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		manager, err := newManagerByName(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Using package manager: %s\n", manager.Name())

		if err := manager.Install(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error installing packages: %v\n", err)
			os.Exit(1)
		}
	},
}
