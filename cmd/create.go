package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/output"
)

var (
	createTags   []string
	createTicket string
	createNoEdit bool
)

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new topic",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := openStore()
		if err != nil {
			return err
		}

		title := ""
		if len(args) > 0 {
			title = strings.TrimSpace(args[0])
		}
		if title == "" {
			title, err = promptTitle()
			if err != nil {
				return err
			}
		}
		if title == "" {
			return fmt.Errorf("title cannot be empty")
		}

		t, err := s.Create(title, createTags, createTicket)
		if err != nil {
			return err
		}

		output.Success("Created: %s [%s]", t.Slug, t.File.Meta.ID)
		output.Info("Path: %s", t.Path)

		if !createNoEdit {
			return openEditor(t.Path)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringArrayVar(&createTags, "tag", nil, "Add a tag (repeatable)")
	createCmd.Flags().StringVar(&createTicket, "ticket", "", "Ticket reference (e.g. PROJ-1234)")
	createCmd.Flags().BoolVar(&createNoEdit, "no-edit", false, "Create file without opening editor")
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
		output.Warnf("$EDITOR not set, falling back to vi. Set with: export EDITOR=<editor>")
	}
	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func promptTitle() (string, error) {
	fmt.Print("Title: ")
	var title string
	_, err := fmt.Scanln(&title)
	return title, err
}
