package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/AmadlaOrg/lay/binary/install"
	"github.com/AmadlaOrg/lay/binary/manifest"
	"github.com/AmadlaOrg/lay/binary/target"
	"github.com/AmadlaOrg/lay/output"
	"github.com/spf13/cobra"
)

// For testability
var (
	newTargetService         = target.NewService
	newInstallService        = install.NewInstallService
	newInstallServiceWithName = install.NewInstallServiceWithName
	newManifestStore         = func() (manifest.Store, error) { return manifest.NewFileStore() }
)

// binaryInstallCmd installs a pre-built binary from GitHub releases or a direct URL
var binaryInstallCmd = &cobra.Command{
	Use:   "install <url|owner/repo>",
	Short: "Install a pre-built binary from GitHub releases or URL",
	Long: `Download and install a pre-built binary or JAR application.

Supports GitHub shorthand (owner/repo) to auto-detect the latest release,
a direct URL to a binary or archive, or a local JAR file.

For JAR files, a launcher script is created that wraps 'java -jar'.
Java must be installed and available in PATH.

Examples:
  lay binary install sharkdp/fd
  lay binary install BurntSushi/ripgrep
  lay binary install --to /opt/bin sharkdp/bat
  lay binary install https://example.com/tool-linux-amd64.tar.gz
  lay binary install ~/Downloads/tika-app-3.2.3.jar
  lay binary install ~/Downloads/tika-app-3.2.3.jar --name tika
  lay binary install apache/tika`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tgt := newTargetService()
		var installer install.Installer
		if binaryNameFlag != "" {
			installer = newInstallServiceWithName(binaryNameFlag)
		} else {
			installer = newInstallService()
		}
		out := StdoutWriter()
		if err := runBinaryInstall(tgt, installer, args[0], binaryToFlag, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runBinaryInstall(tgt target.Target, installer install.Installer, source string, toFlag string, out *output.Writer) error {
	targetDir, err := tgt.Resolve(toFlag)
	if err != nil {
		return fmt.Errorf("resolving target directory: %w", err)
	}

	if err := tgt.Ensure(targetDir); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}

	if flagDryRun {
		out.Info("[dry-run] Would install binary from %s to %s", source, targetDir)
		return nil
	}

	result, err := installer.Install(source, targetDir)
	if err != nil {
		return err
	}

	// Record in manifest
	store, err := newManifestStore()
	if err != nil {
		out.Verbose("Warning: could not open manifest store: %v", err)
		return nil
	}

	m, err := store.Load()
	if err != nil {
		out.Verbose("Warning: could not load manifest: %v", err)
		return nil
	}

	// Remove existing entry if re-installing
	m.Remove(result.BinaryName)

	m.Add(manifest.Entry{
		Name:        result.BinaryName,
		Source:      source,
		Version:     result.Version,
		Path:        result.Path,
		Checksum:    result.Checksum,
		HashAlgo:    result.HashAlgo,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	})

	if err := store.Save(m); err != nil {
		out.Verbose("Warning: could not save manifest: %v", err)
	}

	return nil
}
