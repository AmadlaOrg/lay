package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AmadlaOrg/lay/binary/manifest"
	"github.com/AmadlaOrg/lay/output"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

// binaryListCmd lists installed binaries
var binaryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed binaries tracked by lay",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		out := StdoutWriter()
		store, err := newManifestStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
			return
		}
		if err := runBinaryList(store, os.Stdout, out); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			osExit(1)
		}
	},
}

func runBinaryList(store manifest.Store, w io.Writer, out *output.Writer) error {
	m, err := store.Load()
	if err != nil {
		return err
	}

	if len(m.Entries) == 0 {
		out.Info("No binaries tracked by lay")
		return nil
	}

	out.Result(m.Entries, func(w io.Writer) {
		fmt.Fprintf(w, "Installed binaries: %d\n\n", len(m.Entries))

		table := tablewriter.NewTable(w,
			tablewriter.WithHeaderAlignment(tw.AlignLeft),
			tablewriter.WithRowAlignment(tw.AlignLeft),
			tablewriter.WithRowAutoWrap(tw.WrapNone),
			tablewriter.WithRendition(tw.Rendition{
				Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			}),
		)
		table.Header([]string{"Name", "Version", "Source", "Path"})

		for _, e := range m.Entries {
			table.Append([]string{e.Name, e.Version, e.Source, e.Path})
		}
		table.Render()
	})

	return nil
}
