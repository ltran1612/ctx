package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/output"
	"github.com/user/ctx/internal/store"
)

var (
	listTags     []string
	listArchived bool
	listAll      bool
	listSort     string
	listFormat   string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List topics",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := openStore()
		if err != nil {
			return err
		}

		includeActive := !listArchived
		includeArchive := listArchived || listAll
		if listAll {
			includeActive = true
		}

		topics, err := s.All(includeActive, includeArchive)
		if err != nil {
			return err
		}

		// filter by tags
		if len(listTags) > 0 {
			topics = filterByTags(topics, listTags)
		}

		if len(topics) == 0 {
			output.Info("No topics found. Create one with: ctx create \"My first topic\"")
			return nil
		}

		// sort
		sortTopics(topics, listSort)

		switch listFormat {
		case "ids":
			for _, t := range topics {
				fmt.Println(t.File.Meta.ID)
			}
		case "slugs":
			for _, t := range topics {
				fmt.Println(t.Slug)
			}
		default:
			printTopicTable(topics, listAll || listArchived)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringArrayVar(&listTags, "tag", nil, "Filter by tag (repeatable, AND logic)")
	listCmd.Flags().BoolVar(&listArchived, "archived", false, "Show archived topics only")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Show active and archived topics")
	listCmd.Flags().StringVar(&listSort, "sort", "updated", "Sort by: updated, created, title")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format: table, ids, slugs")
}

func filterByTags(topics []*store.Topic, tags []string) []*store.Topic {
	var out []*store.Topic
	for _, t := range topics {
		if hasAllTags(t.File.Meta.Tags, tags) {
			out = append(out, t)
		}
	}
	return out
}

func hasAllTags(have, want []string) bool {
	for _, w := range want {
		found := false
		for _, h := range have {
			if h == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func sortTopics(topics []*store.Topic, by string) {
	sort.Slice(topics, func(i, j int) bool {
		mi, mj := topics[i].File.Meta, topics[j].File.Meta
		switch by {
		case "created":
			return mi.Created.After(mj.Created)
		case "title":
			return mi.Title < mj.Title
		default: // updated
			return mi.Updated.After(mj.Updated)
		}
	})
}

func printTopicTable(topics []*store.Topic, showStatus bool) {
	headers := []string{"ID", "SLUG", "TITLE", "UPDATED", "TAGS"}
	if showStatus {
		headers = []string{"ID", "SLUG", "TITLE", "STATUS", "UPDATED", "TAGS"}
	}
	rows := make([][]string, len(topics))
	for i, t := range topics {
		m := t.File.Meta
		tags := strings.Join(m.Tags, ", ")
		updated := m.Updated.Format("2006-01-02")
		if showStatus {
			rows[i] = []string{m.ID, t.Slug, m.Title, m.Status, updated, tags}
		} else {
			rows[i] = []string{m.ID, t.Slug, m.Title, updated, tags}
		}
	}
	output.Table(headers, rows)
}
