package kv

import (
	"github.com/spf13/cobra"
)

var namespaceAccountID string

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Manage KV namespaces",
}

func init() {
	namespaceCmd.PersistentFlags().StringVar(&namespaceAccountID, "account-id", "", "The account ID")
	KVCmd.AddCommand(namespaceCmd)
}
