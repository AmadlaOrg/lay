package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AmadlaOrg/lay/binary/compile"
	"github.com/AmadlaOrg/lay/binary/install"
	"github.com/AmadlaOrg/lay/binary/target"
	"github.com/spf13/cobra"
)

var buildSystemFlag string

// For testability
var (
	newCompileTargetService = target.NewService
	newCompileDetector      = compile.NewDetectorService
	newBuildSystemByName    = compile.NewBuildSystemByName
)

// binaryCompileCmd compiles a binary from source
var binaryCompileCmd = &cobra.Command{
	Use:   "compile <url|path>",
	Short: "Compile a binary from source",
	Long: `Clone or use local source, detect the build system, compile, and install.

Supported build systems: autotools, cmake, meson, makefile, cargo, golang.
Use --build-system to override auto-detection.

Examples:
  lay binary compile .
  lay binary compile --build-system golang .
  lay binary compile https://github.com/user/repo
  lay binary compile user/repo`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tgt := newCompileTargetService()

		targetDir, err := tgt.Resolve(binaryToFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving target directory: %v\n", err)
			os.Exit(1)
		}

		if err := tgt.Ensure(targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating target directory: %v\n", err)
			os.Exit(1)
		}

		// Clone or resolve local source
		srcDir, cleanup, err := compile.CloneOrDownload(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()

		// Detect build system
		detector := newCompileDetector()
		buildSystem, err := detector.Detect(srcDir, buildSystemFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Detected build system: %s\n", buildSystem)

		// Get build system implementation
		bs, err := newBuildSystemByName(buildSystem)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Build into a temporary prefix directory
		tmpBuildDir, err := os.MkdirTemp("", "lay-build-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating build directory: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tmpBuildDir)

		if err := bs.Build(srcDir, tmpBuildDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Find the built binary and copy to target
		binaryPath, err := install.FindBinary(tmpBuildDir, filepath.Base(srcDir))
		if err != nil {
			// For cargo/golang, the binary may be in the source dir itself
			binaryPath, err = install.FindBinary(srcDir, filepath.Base(srcDir))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error finding built binary: %v\n", err)
				os.Exit(1)
			}
		}

		destPath := filepath.Join(targetDir, filepath.Base(binaryPath))
		if err := copyBinaryFile(binaryPath, destPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error copying binary: %v\n", err)
			os.Exit(1)
		}

		if err := os.Chmod(destPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting permissions: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Installed %s to %s\n", filepath.Base(binaryPath), destPath)
	},
}

func init() {
	binaryCompileCmd.Flags().StringVar(&buildSystemFlag, "build-system", "", "Override build system detection (autotools, cmake, meson, makefile, cargo, golang)")
}

func copyBinaryFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}
