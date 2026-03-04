package cmd

import "github.com/spf13/cobra"

var binaryToFlag string
var binaryNameFlag string

// BinaryCmd is the parent command for binary management operations
var BinaryCmd = &cobra.Command{
	Use:   "binary",
	Short: "Install or compile standalone binaries",
	Long: `Install pre-built binaries from GitHub releases or compile from source.

Override the install directory with --to flag or LAY_BINARY_PATH env var.
Default install directory: ~/.local/bin

Examples:
  lay binary install sharkdp/fd
  lay binary install --to /opt/bin sharkdp/fd
  lay binary compile --build-system golang .
  lay binary compile https://github.com/user/repo`,
}

func init() {
	BinaryCmd.PersistentFlags().StringVar(&binaryToFlag, "to", "", "Override install directory (default: ~/.local/bin)")
	BinaryCmd.PersistentFlags().StringVar(&binaryNameFlag, "name", "", "Override the installed command name")
	BinaryCmd.AddCommand(binaryInstallCmd)
	BinaryCmd.AddCommand(binaryRemoveCmd)
	BinaryCmd.AddCommand(binaryCompileCmd)
	BinaryCmd.AddCommand(binaryListCmd)
	BinaryCmd.AddCommand(binaryUpdateCmd)
}
