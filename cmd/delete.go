package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/output"
)

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <topic>",
	Short: "Permanently delete a topic",
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

		if !deleteConfirm {
			fmt.Printf("Delete %q [%s]? This cannot be undone. [y/N] ", t.File.Meta.Title, t.File.Meta.ID)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.ToLower(strings.TrimSpace(answer))
			if answer != "y" && answer != "yes" {
				output.Info("Aborted.")
				return nil
			}
		}

		if err := s.Delete(t); err != nil {
			return err
		}
		output.Success("Deleted: %s [%s]", t.Slug, t.File.Meta.ID)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteConfirm, "confirm", false, "Skip confirmation prompt")
}
