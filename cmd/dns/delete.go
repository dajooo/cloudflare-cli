package dns

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <zone> <record_id>",
	Short: "Deletes a DNS record",
	Args:  cobra.ExactArgs(2),
	Run: executor.NewBuilder[*cf.Client, *RecordInformation]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Deleting DNS record", deleteDnsRecord).
		Display(func(deletedRecord *RecordInformation, duration time.Duration, err error) {
			if err != nil {
				fmt.Println(ui.ErrorMessage("Error deleting DNS record", err))
				return
			}
			fmt.Println(ui.Success(fmt.Sprintf("Successfully deleted DNS record %s.%s (%s)", deletedRecord.RecordName, deletedRecord.ZoneName, deletedRecord.RecordID)))
		}).
		Build().
		CobraRun(),
}

func deleteDnsRecord(client *cf.Client, _ *cobra.Command, args []string) (*RecordInformation, error) {
	zoneIdentifier := args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %s", err)
	}

	recordName := args[1]

	recordID, err := cloudflare.LookupDNSRecordID(client, zoneID, zoneName, recordName)
	if err != nil {
		return nil, fmt.Errorf("error finding record: %s", err)
	}

	_, err = client.DNS.Records.Delete(context.Background(), recordID, dns.RecordDeleteParams{
		ZoneID: cf.F(zoneID),
	})

	return &RecordInformation{
		ZoneID:     zoneID,
		ZoneName:   zoneName,
		RecordID:   recordID,
		RecordName: recordName,
	}, nil
}

func init() {
	DnsCmd.AddCommand(deleteCmd)
}
