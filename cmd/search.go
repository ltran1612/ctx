package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/output"
)

var (
	searchFullText bool
	searchLimit    int
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Fuzzy search topics by title",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := openStore()
		if err != nil {
			return err
		}

		results, err := s.Search(args[0], searchFullText, searchLimit)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			output.Info("No topics matched %q", args[0])
			return nil
		}

		headers := []string{"SCORE", "ID", "SLUG", "TITLE", "TAGS"}
		rows := make([][]string, len(results))
		for i, r := range results {
			m := r.Topic.File.Meta
			status := ""
			if r.Archived {
				status = " (archived)"
			}
			rows[i] = []string{
				fmt.Sprintf("%d", r.Score),
				m.ID,
				r.Topic.Slug + status,
				m.Title,
				strings.Join(m.Tags, ", "),
			}
		}
		output.Table(headers, rows)
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolVar(&searchFullText, "full-text", false, "Search body content in addition to title")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Maximum number of results")
}
