package cmd

import (
	"encoding/json"
	"os"

	"github.com/gobuffalo/buffalo-plugins/plugins"
	"github.com/spf13/cobra"
)

// availableCmd represents the available command
var availableCmd = &cobra.Command{
	Use:   "available",
	Short: "a list of available buffalo plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		plugs := plugins.Commands{
			{
				Name:           testCmd.Use,
				BuffaloCommand: "root",
				Description:    testCmd.Short,
				Aliases:        testCmd.Aliases,
			},
		}

		return json.NewEncoder(os.Stdout).Encode(plugs)
	},
}

func init() {
	rootCmd.AddCommand(availableCmd)
}
