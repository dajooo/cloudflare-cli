package dns

import (
	"github.com/spf13/cobra"
)

type RecordInformation struct {
	ZoneID     string `json:"zone_id"`
	ZoneName   string `json:"zone_name"`
	RecordID   string `json:"record_id"`
	RecordName string `json:"record_name"`
}

var DnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Manage Cloudflare DNS records",
}
