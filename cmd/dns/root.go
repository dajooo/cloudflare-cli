package dns

import (
	"github.com/spf13/cobra"
)

var DnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Manage Cloudflare DNS records",
}
