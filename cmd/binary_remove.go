package cmd

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/lay/binary/jar"
	"github.com/AmadlaOrg/lay/binary/manifest"
	"github.com/AmadlaOrg/lay/output"
	"github.com/spf13/cobra"
)

// For testability
var (
	binaryRemoveManifestStore = func() (manifest.Store, error) { return manifest.NewFileStore() }
	jarIsLauncherScript       = jar.IsLauncherScript
	jarParseLauncherScript    = jar.ParseLauncherScript
	binaryRemoveOsRemove      = os.Remove
)

// binaryRemoveCmd removes a lay-managed binary
var binaryRemoveCmd = &cobra.Command{
	Use:   "remove <name>...",
	Short: "Remove installed binaries",
	Long: `Remove binaries previously installed with 'lay binary install'.

Removes the binary file from disk, the JAR file (if applicable),
and the entry from the manifest.

Examples:
  lay binary remove fd
  lay binary remove fd bat ripgrep
  lay binary remove --dry-run tika-app`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out := StdoutWriter()
		if err := runBinaryRemove(args, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runBinaryRemove(names []string, out *output.Writer) error {
	store, err := binaryRemoveManifestStore()
	if err != nil {
		return fmt.Errorf("opening manifest: %w", err)
	}

	m, err := store.Load()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	for _, name := range names {
		entry := m.Find(name)
		if entry == nil {
			out.Info("Not found in manifest: %s", name)
			continue
		}

		if flagDryRun {
			out.Info("[dry-run] Would remove: %s (%s)", name, entry.Path)
			continue
		}

		// Check if it's a JAR launcher script
		isJar, _ := jarIsLauncherScript(entry.Path)
		if isJar {
			jarPath, err := jarParseLauncherScript(entry.Path)
			if err == nil {
				if err := binaryRemoveOsRemove(jarPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("removing JAR %s: %w", jarPath, err)
				}
			}
		}

		// Remove the binary/launcher script
		if err := binaryRemoveOsRemove(entry.Path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing binary %s: %w", entry.Path, err)
		}

		m.Remove(name)
		out.Info("Removed: %s", name)
	}

	if !flagDryRun {
		if err := store.Save(m); err != nil {
			return fmt.Errorf("saving manifest: %w", err)
		}
	}

	return nil
}
