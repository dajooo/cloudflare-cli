package cmd

import (
	"github.com/dajooo/cloudflare-cli/cmd/dns"
)

func init() {
	rootCmd.AddCommand(dns.DnsCmd)
}
