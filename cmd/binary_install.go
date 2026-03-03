package cmd

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/lay/binary/install"
	"github.com/AmadlaOrg/lay/binary/target"
	"github.com/spf13/cobra"
)

// For testability
var (
	newTargetService  = target.NewService
	newInstallService = install.NewInstallService
)

// binaryInstallCmd installs a pre-built binary from GitHub releases or a direct URL
var binaryInstallCmd = &cobra.Command{
	Use:   "install <url|owner/repo>",
	Short: "Install a pre-built binary from GitHub releases or URL",
	Long: `Download and install a pre-built binary.

Supports GitHub shorthand (owner/repo) to auto-detect the latest release,
or a direct URL to a binary or archive.

Examples:
  lay binary install sharkdp/fd
  lay binary install BurntSushi/ripgrep
  lay binary install --to /opt/bin sharkdp/bat
  lay binary install https://example.com/tool-linux-amd64.tar.gz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tgt := newTargetService()

		targetDir, err := tgt.Resolve(binaryToFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving target directory: %v\n", err)
			os.Exit(1)
		}

		if err := tgt.Ensure(targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating target directory: %v\n", err)
			os.Exit(1)
		}

		installer := newInstallService()
		if err := installer.Install(args[0], targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}
