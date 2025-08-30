package cmd

import (
	"github.com/dajooo/cloudflare-cli/cmd/zone"
)

func init() {
	rootCmd.AddCommand(zone.ZoneCmd)
}
