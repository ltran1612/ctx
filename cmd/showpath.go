package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var showPathCmd = &cobra.Command{
	Use:   "show-path <topic>",
	Short: "Print the filesystem path to a topic's context.md",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := openStore()
		if err != nil {
			return err
		}
		t, _, err := s.Resolve(args[0])
		if err != nil {
			return err
		}
		fmt.Println(t.Path)
		return nil
	},
}
