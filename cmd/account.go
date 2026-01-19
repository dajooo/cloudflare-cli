package cmd

import (
	"dario.lol/cf/cmd/account"
)

func init() {
	rootCmd.AddCommand(account.AccountCmd)
}
