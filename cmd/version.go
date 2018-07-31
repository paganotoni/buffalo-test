package cmd

import (
	"fmt"

	"github.com/paganotoni/buffalo-test/test"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "current version of test",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("test", test.Version)
		return nil
	},
}

func init() {
	testCmd.AddCommand(versionCmd)
}
