package dns

import (
	"context"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/dajooo/cloudflare-cli/internal/cloudflare"
	"github.com/dajooo/cloudflare-cli/internal/ui"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <zone> <record_id>",
	Short: "Deletes a DNS record",
	Args:  cobra.ExactArgs(2),
	Run:   executeDnsDelete,
}

func init() {
	DnsCmd.AddCommand(deleteCmd)
}

func executeDnsDelete(cmd *cobra.Command, args []string) {
	client, err := cloudflare.NewClient()
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error loading config", err))
		return
	}

	zoneIdentifier := args[0]
	zoneID, err := cloudflare.ZoneIDByName(client, zoneIdentifier)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error finding zone", err))
		return
	}

	recordName := args[1]

	recordID, err := cloudflare.DNSRecordIDByName(client, zoneID, zoneIdentifier, recordName)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error finding record", err))
		return
	}

	_, err = client.DNS.Records.Delete(context.Background(), recordID, dns.RecordDeleteParams{
		ZoneID: cf.F(zoneID),
	})
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error deleting DNS record", err))
		return
	}

	fmt.Println(ui.Success(fmt.Sprintf("Successfully deleted DNS record %s.%s", recordName, zoneIdentifier)))
}
