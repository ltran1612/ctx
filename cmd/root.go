package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var noColor bool

var rootCmd = &cobra.Command{
	Use:   "ctx",
	Short: "ctx — project-scoped context management for humans and AI agents",
	Long: `ctx keeps context across tickets and projects.
Topics are stored as markdown files in a .ctx/ directory in the current project.

Run 'ctx init' to set up a new context store.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable ANSI color output")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if noColor {
			os.Setenv("NO_COLOR", "1")
		}
	}

	rootCmd.AddCommand(
		initCmd,
		createCmd,
		viewCmd,
		editCmd,
		listCmd,
		searchCmd,
		archiveCmd,
		restoreCmd,
		deleteCmd,
		showPathCmd,
	)
}
