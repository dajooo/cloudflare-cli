package kv

import (
	"github.com/spf13/cobra"
)

var KVCmd = &cobra.Command{
	Use:   "kv",
	Short: "Manage KV namespaces and keys",
}
