package kv

import (
	"github.com/spf13/cobra"
)

var keyAccountID string
var namespaceID string

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage KV keys",
}

func init() {
	keyCmd.PersistentFlags().StringVar(&keyAccountID, "account-id", "", "The account ID")
	keyCmd.PersistentFlags().StringVar(&namespaceID, "namespace-id", "", "The namespace ID")

	KVCmd.AddCommand(keyCmd)
}
