package cmd

import (
	"dario.lol/cf/cmd/d1"
)

func init() {
	rootCmd.AddCommand(d1.D1Cmd)
}
