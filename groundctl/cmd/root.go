package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var outputFormat string

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "groundctl",
	Short: "CLI for Harbor Satellite Ground Control",
	Long: `groundctl is a command-line tool for managing Harbor Satellite fleets
through the Ground Control API.

It provides commands for satellite lifecycle management, group and config
operations, and a GitOps-style 'apply -f' workflow for declarative fleet
configuration.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table or json")
}
