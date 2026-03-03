package cmd

import "github.com/spf13/cobra"

// PackageCmd is the parent command for package management operations
var PackageCmd = &cobra.Command{
	Use:   "package",
	Short: "Manage system packages (install, search)",
	Long: `Manage system packages using the auto-detected package manager.

Override the package manager with --manager flag or LAY_PACKAGE_MANAGER env var.

Examples:
  lay package install curl wget
  lay package search nodejs
  lay package --manager dnf install vim`,
}

func init() {
	PackageCmd.PersistentFlags().StringVar(&managerFlag, "manager", "", "Override package manager (apt, dnf, yum, pacman, zypper, apk, nix, snap, flatpak, dpkg, rpm)")
	PackageCmd.AddCommand(installCmd)
	PackageCmd.AddCommand(searchCmd)
}
