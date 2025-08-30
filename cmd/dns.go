package cmd

import (
	"dario.lol/cf/cmd/dns"
)

func init() {
	rootCmd.AddCommand(dns.DnsCmd)
}
