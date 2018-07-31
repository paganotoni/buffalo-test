package cmd

import (
	"fmt"

	"github.com/paganotoni/buffalo-tester/tester"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "current version of test",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("tester", tester.Version)
		return nil
	},
}

func init() {
	testCmd.AddCommand(versionCmd)
}
