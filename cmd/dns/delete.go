package dns

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
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
		Invalidates(func(cmd *cobra.Command, args []string, result *RecordInformation) []string {
			return []string{fmt.Sprintf("zone:%s", result.ZoneID)}
		}).
		Display(printDeleteDnsResult).
		Build().
		CobraRun(),
}

func init() {
	DnsCmd.AddCommand(deleteCmd)
}

func deleteDnsRecord(client *cf.Client, _ *cobra.Command, args []string, progress chan<- string) (*RecordInformation, error) {
	zoneIdentifier := args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %s", err)
	}

	recordIdentifier := args[1]
	recordID, err := cloudflare.LookupDNSRecordID(client, zoneID, zoneName, recordIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding record: %s", err)
	}

	_, err = client.DNS.Records.Delete(context.Background(), recordID, dns.RecordDeleteParams{
		ZoneID: cf.F(zoneID),
	})
	if err != nil {
		return nil, err
	}

	return &RecordInformation{
		ZoneID:     zoneID,
		ZoneName:   zoneName,
		RecordID:   recordID,
		RecordName: recordIdentifier,
	}, nil
}

func printDeleteDnsResult(deletedRecord *RecordInformation, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error deleting DNS record", err).Display()
		return
	}
	rb.FooterSuccess("Successfully deleted DNS record %s (%s) in zone %s %s", deletedRecord.RecordName, deletedRecord.RecordID, deletedRecord.ZoneName, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
