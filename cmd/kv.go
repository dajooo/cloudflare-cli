package cmd

import (
	"dario.lol/cf/cmd/kv"
)

func init() {
	rootCmd.AddCommand(kv.KVCmd)
}
