package cmd

import (
	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/frontmatter"
	"github.com/user/ctx/internal/output"
)

var (
	editAppend      string
	editPrependNote string
	editSetTitle    string
	editAddTags     []string
	editRemoveTags  []string
	editSetTicket   string
)

var editCmd = &cobra.Command{
	Use:   "edit <topic>",
	Short: "Edit a topic (opens $EDITOR, or use flags for non-interactive edits)",
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

		// Non-interactive edits
		modified := false

		if editSetTitle != "" {
			t.File.Meta.Title = editSetTitle
			modified = true
		}
		if editSetTicket != "" {
			t.File.Meta.Ticket = editSetTicket
			modified = true
		}
		for _, tag := range editAddTags {
			if !containsTag(t.File.Meta.Tags, tag) {
				t.File.Meta.Tags = append(t.File.Meta.Tags, tag)
			}
			modified = true
		}
		for _, tag := range editRemoveTags {
			t.File.Meta.Tags = removeTags(t.File.Meta.Tags, tag)
			modified = true
		}
		if editAppend != "" {
			t.File.Body = frontmatter.AppendToSection(t.File.Body, "Notes", editAppend)
			modified = true
		}
		if editPrependNote != "" {
			t.File.Body = frontmatter.PrependToSection(t.File.Body, "Notes", editPrependNote)
			modified = true
		}

		if modified {
			if err := s.Save(t); err != nil {
				return err
			}
			output.Success("Updated: %s [%s]", t.Slug, t.File.Meta.ID)
			return nil
		}

		// Interactive: open editor
		if err := openEditor(t.Path); err != nil {
			return err
		}
		// Reload from disk (editor may have changed content) then stamp Updated
		if err := s.Reload(t); err != nil {
			return err
		}
		if err := s.Save(t); err != nil {
			return err
		}
		output.Success("Updated: %s [%s]", t.Slug, t.File.Meta.ID)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVar(&editAppend, "append", "", "Append text to the Notes section")
	editCmd.Flags().StringVar(&editPrependNote, "prepend-note", "", "Prepend a timestamped note to the Notes section")
	editCmd.Flags().StringVar(&editSetTitle, "set-title", "", "Update the title (slug unchanged)")
	editCmd.Flags().StringArrayVar(&editAddTags, "add-tag", nil, "Add a tag")
	editCmd.Flags().StringArrayVar(&editRemoveTags, "remove-tag", nil, "Remove a tag")
	editCmd.Flags().StringVar(&editSetTicket, "set-ticket", "", "Update ticket reference")
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func removeTags(tags []string, remove string) []string {
	var out []string
	for _, t := range tags {
		if t != remove {
			out = append(out, t)
		}
	}
	return out
}
