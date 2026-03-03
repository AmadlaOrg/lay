package cmd

import (
	"fmt"
	"os"

	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// searchCmd searches for packages using the detected system package manager
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for packages using the system package manager",
	Args:  cobra.ExactArgs(1),
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

		results, err := manager.Search(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching packages: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Printf("No packages found for '%s' (using %s)\n", args[0], manager.Name())
			return
		}

		renderSearchResults(manager.Name(), results)
	},
}

func renderSearchResults(managerName string, results []pm.SearchResult) {
	fmt.Printf("Search results (%s):\n\n", managerName)

	table := tablewriter.NewWriter(os.Stdout)

	// Check if any results have versions
	hasVersions := false
	for _, r := range results {
		if r.Version != "" {
			hasVersions = true
			break
		}
	}

	if hasVersions {
		table.SetHeader([]string{"Name", "Version", "Description"})
	} else {
		table.SetHeader([]string{"Name", "Description"})
	}

	table.SetAutoWrapText(false)
	table.SetBorder(false)
	table.SetColumnSeparator(" ")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

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
