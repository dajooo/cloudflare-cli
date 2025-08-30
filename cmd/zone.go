package cmd

import (
	"dario.lol/cf/cmd/zone"
)

func init() {
	rootCmd.AddCommand(zone.ZoneCmd)
}
