package cmd

import (
	"dario.lol/cf/cmd/r2"
)

func init() {
	rootCmd.AddCommand(r2.R2Cmd)
}
