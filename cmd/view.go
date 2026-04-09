package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/frontmatter"
	"github.com/user/ctx/internal/output"
)

var (
	viewSection string
	viewRaw     bool
)

var viewCmd = &cobra.Command{
	Use:   "view <topic>",
	Short: "View a topic's context file",
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
			output.Warnf("note: this topic is archived")
		}

		if viewSection != "" {
			content, ok := frontmatter.Section(t.File.Body, viewSection)
			if !ok {
				return fmt.Errorf("section %q not found", viewSection)
			}
			fmt.Println(content)
			return nil
		}

		if viewRaw {
			data, err := os.ReadFile(t.Path)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
			return nil
		}

		printTopic(t.File)
		return nil
	},
}

func init() {
	viewCmd.Flags().StringVar(&viewSection, "section", "", "Print only a specific section (e.g. \"Next Steps\")")
	viewCmd.Flags().BoolVar(&viewRaw, "raw", false, "Print raw markdown without formatting")
}

func printTopic(f *frontmatter.File) {
	m := f.Meta
	// Print frontmatter as a clean header block
	output.Bold.Printf("%-10s %s\n", "Title:", m.Title)
	output.Dim.Printf("%-10s %s\n", "ID:", m.ID)
	output.Dim.Printf("%-10s %s\n", "Slug:", m.Slug)
	output.Dim.Printf("%-10s %s\n", "Status:", m.Status)
	output.Dim.Printf("%-10s %s\n", "Updated:", m.Updated.Format("2006-01-02 15:04"))
	if m.Ticket != "" {
		output.Dim.Printf("%-10s %s\n", "Ticket:", m.Ticket)
	}
	if len(m.Tags) > 0 {
		output.Dim.Printf("%-10s %s\n", "Tags:", strings.Join(m.Tags, ", "))
	}
	fmt.Println(strings.Repeat("─", 60))
	output.FprintMarkdown(os.Stdout, f.Body)
}
