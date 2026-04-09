package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/ctx/internal/output"
	"github.com/user/ctx/internal/store"
)

var initGitignore bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a context store in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		if _, err := store.Init(dir); err != nil {
			return fmt.Errorf("init failed: %w", err)
		}

		output.Success("Initialized context store at .ctx/")

		if initGitignore {
			if err := appendGitignore(dir); err != nil {
				output.Warnf("could not update .gitignore: %v", err)
			} else {
				output.Info("Added .ctx/ to .gitignore")
			}
		}
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initGitignore, "gitignore", false, "Append .ctx/ to .gitignore")
}

func appendGitignore(dir string) error {
	path := filepath.Join(dir, ".gitignore")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, ".ctx/")
	return err
}
