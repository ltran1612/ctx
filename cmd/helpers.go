package cmd

import (
	"github.com/user/ctx/internal/store"
)

// openStore opens the context store from ~/.ctx/, creating it if needed.
func openStore() (*store.Store, error) {
	return store.Open()
}
