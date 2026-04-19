package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AmadlaOrg/lay/output"
	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

// searchCmd searches for packages using the detected system package manager
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for packages using the system package manager",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		detector := newDetector()
		out := StdoutWriter()
		if err := runSearch(detector, newManagerByName, managerFlag, args[0], os.Stdout, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runSearch(detector pm.Detector, managerFn func(string) (pm.Manager, error), override string, query string, w io.Writer, out *output.Writer) error {
	name, err := detector.Detect(override)
	if err != nil {
		return err
	}

	manager, err := managerFn(name)
	if err != nil {
		return err
	}

	results, err := manager.Search(query)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		out.Info("No packages found for '%s' (using %s)", query, manager.Name())
		return nil
	}

	out.Result(results, func(w io.Writer) {
		renderSearchResults(w, manager.Name(), results)
	})
	return nil
}

func renderSearchResults(w io.Writer, managerName string, results []pm.SearchResult) {
	fmt.Fprintf(w, "Search results (%s):\n\n", managerName)

	table := tablewriter.NewTable(w,
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
		}),
	)

	// Check if any results have versions
	hasVersions := false
	for _, r := range results {
		if r.Version != "" {
			hasVersions = true
			break
		}
	}

	if hasVersions {
		table.Header([]string{"Name", "Version", "Description"})
	} else {
		table.Header([]string{"Name", "Description"})
	}

	for _, r := range results {
		desc := r.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		if hasVersions {
			table.Append([]string{r.Name, r.Version, desc})
		} else {
			table.Append([]string{r.Name, desc})
		}
	}

	table.Render()
}
