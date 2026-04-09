package cmd

import (
	"os"

	"github.com/user/ctx/internal/store"
)

// openStore opens the context store from the current working directory.
func openStore() (*store.Store, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return store.Open(dir)
}
