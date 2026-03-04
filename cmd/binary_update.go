package cmd

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/lay/binary/install"
	"github.com/AmadlaOrg/lay/binary/manifest"
	"github.com/AmadlaOrg/lay/output"
	"github.com/spf13/cobra"
)

// binaryUpdateCmd updates installed binaries to the latest version
var binaryUpdateCmd = &cobra.Command{
	Use:   "update [name...]",
	Short: "Update installed binaries to the latest version",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		out := StdoutWriter()
		store, err := newManifestStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
			return
		}
		installer := newInstallService()
		if err := runBinaryUpdate(store, installer, args, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runBinaryUpdate(store manifest.Store, installer install.Installer, names []string, out *output.Writer) error {
	m, err := store.Load()
	if err != nil {
		return err
	}

	if len(m.Entries) == 0 {
		out.Info("No binaries tracked by lay")
		return nil
	}

	// Determine which entries to update
	var entries []manifest.Entry
	if len(names) == 0 {
		entries = m.Entries
	} else {
		for _, name := range names {
			e := m.Find(name)
			if e == nil {
				return fmt.Errorf("binary not found in manifest: %s", name)
			}
			entries = append(entries, *e)
		}
	}

	for _, e := range entries {
		out.Info("Checking %s (current: %s)...", e.Name, e.Version)

		if flagDryRun {
			out.Info("[dry-run] Would update %s from source %s", e.Name, e.Source)
			continue
		}

		// Re-install from the original source to the same directory
		targetDir := e.Path
		if lastSlash := lastIndexByte(targetDir, '/'); lastSlash >= 0 {
			targetDir = targetDir[:lastSlash]
		}

		result, err := installer.Install(e.Source, targetDir)
		if err != nil {
			out.Info("Failed to update %s: %v", e.Name, err)
			continue
		}

		if result.Version == e.Version {
			out.Info("%s is already at latest version (%s)", e.Name, e.Version)
			continue
		}

		m.Update(e.Name, result.Version, result.Checksum)
		out.Info("Updated %s: %s -> %s", e.Name, e.Version, result.Version)
	}

	return store.Save(m)
}

func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
