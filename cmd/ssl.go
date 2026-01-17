package cmd

import (
	"dario.lol/cf/cmd/ssl"
)

func init() {
	rootCmd.AddCommand(ssl.SslCmd)
}
