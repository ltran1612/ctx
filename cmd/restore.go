package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/output"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <topic>",
	Short: "Restore an archived topic back to active",
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
		if !archived {
			return fmt.Errorf("topic %q is not archived", args[0])
		}

		if err := s.Restore(t); err != nil {
			return err
		}
		output.Success("Restored: %s [%s]", t.Slug, t.File.Meta.ID)
		return nil
	},
}
