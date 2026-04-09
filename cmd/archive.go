package cmd

import (
	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/output"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <topic>",
	Short: "Archive a topic (moves it to .ctx/archive/)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := openStore()
		if err != nil {
			return err
		}

		t, archived, err := s.Resolve(args[0])
		if err != nil {
			return err
		}
		if archived {
			output.Info("Topic %q is already archived", args[0])
			return nil
		}

		if err := s.Archive(t); err != nil {
			return err
		}
		output.Success("Archived: %s [%s]", t.Slug, t.File.Meta.ID)
		return nil
	},
}
