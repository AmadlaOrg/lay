package cmd

import (
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
)

var SettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "List the paths and other environment variables for HERY",
	Run: func(cmd *cobra.Command, args []string) {
		/*storageService := storage.NewStorageService()
		heryPath, err := storageService.Main()
		if err != nil {
			log.Fatal(err)
		}*/

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Setting", "Value"})
		table.Append([]string{"Collections path", "..."})

		/*envList, err := env.List()
		if err != nil {
			log.Fatal(err)
		}

		for _, varName := range envList {
			val := os.Getenv(varName)
			table.Append([]string{varName, val})
		}*/

		table.Render()
	},
}
