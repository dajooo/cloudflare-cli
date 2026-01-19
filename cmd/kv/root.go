package kv

import (
	"dario.lol/cf/internal/flags"
	"github.com/spf13/cobra"
)

var KVCmd = &cobra.Command{
	Use:   "kv",
	Short: "Manage KV namespaces and keys",
}

func init() {
	flags.RegisterAccountID(KVCmd)
}
