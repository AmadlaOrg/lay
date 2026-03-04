package cmd

import (
	"fmt"
	"os"

	ct "github.com/AmadlaOrg/lay/container"
	"github.com/AmadlaOrg/lay/output"
	"github.com/spf13/cobra"
)

// For testability
var (
	newContainerDetector = ct.NewDetectorService
	newRuntimeByName     = ct.NewRuntimeByName
)

// ContainerCmd runs container commands using the detected runtime (podman or docker)
var ContainerCmd = &cobra.Command{
	Use:   "container [args...]",
	Short: "Run container commands using docker or podman",
	Long: `Run container commands using the auto-detected container runtime.
Podman is preferred over Docker when both are available.

Override the runtime with --runtime flag or LAY_CONTAINER_RUNTIME env var.

Examples:
  lay container --version
  lay container run --rm alpine echo hello
  lay container --runtime docker ps
  lay container docker build -t myapp .
  lay container podman run -it ubuntu`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Manually extract --runtime flag from args
		runtimeOverride, remaining := extractRuntimeFlag(args)

		// Check for help flags
		for _, a := range remaining {
			if a == "-h" || a == "--help" {
				_ = cmd.Help()
				return
			}
		}

		if len(remaining) == 0 {
			_ = cmd.Help()
			return
		}

		// Check if first arg is a subcommand name
		switch remaining[0] {
		case "docker":
			execWithRuntime("docker", remaining[1:])
		case "podman":
			execWithRuntime("podman", remaining[1:])
		default:
			execWithRuntime(runtimeOverride, remaining)
		}
	},
}

// extractRuntimeFlag parses --runtime <value> or --runtime=<value> from args,
// returning the runtime value and remaining args
func extractRuntimeFlag(args []string) (string, []string) {
	var runtime string
	var remaining []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--runtime" && i+1 < len(args):
			runtime = args[i+1]
			i++ // skip the value
		case len(a) > 10 && a[:10] == "--runtime=":
			runtime = a[10:]
		default:
			remaining = append(remaining, a)
		}
	}

	return runtime, remaining
}

func execWithRuntime(runtimeOverride string, args []string) {
	detector := newContainerDetector()
	out := StdoutWriter()
	if err := runExecWithRuntime(detector, newRuntimeByName, runtimeOverride, args, out); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		osExit(1)
	}
}

func runExecWithRuntime(detector ct.Detector, runtimeFn func(string) (ct.Runtime, error), override string, args []string, out *output.Writer) error {
	name, err := detector.Detect(override)
	if err != nil {
		return err
	}

	runtime, err := runtimeFn(name)
	if err != nil {
		return err
	}

	out.Info("Using container runtime: %s", runtime.Name())

	if flagDryRun {
		out.Info("[dry-run] Would execute: %s %v", runtime.Name(), args)
		return nil
	}

	return runtime.Exec(args)
}
