package dns

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/spf13/cobra"
)

var deletedRecordKey = executor.NewKey[*RecordInformation]("deletedRecord")

var deleteCmd = &cobra.Command{
	Use:   "delete <zone> <record>",
	Short: "Deletes a DNS record",
	Args:  cobra.ExactArgs(2),
	Run: executor.New().
		WithClient().
		Step(executor.NewStep(deletedRecordKey, "Deleting DNS record").Func(deleteDnsRecord)).
		Invalidates(func(ctx *executor.Context) []string {
			record := executor.Get(ctx, deletedRecordKey)
			if record != nil {
				return []string{fmt.Sprintf("zone:%s", record.ZoneID)}
			}
			return nil
		}).
		Display(printDeleteDnsResult).
		Run(),
}

func init() {
	DnsCmd.AddCommand(deleteCmd)
}

func deleteDnsRecord(ctx *executor.Context, _ chan<- string) (*RecordInformation, error) {
	zoneIdentifier := ctx.Args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %s", err)
	}

	recordIdentifier := ctx.Args[1]
	recordID, err := cloudflare.LookupDNSRecordID(ctx.Client, zoneID, zoneName, recordIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding record: %s", err)
	}

	_, err = ctx.Client.DNS.Records.Delete(context.Background(), recordID, dns.RecordDeleteParams{
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

func printDeleteDnsResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error deleting DNS record", ctx.Error).Display()
		return
	}
	deletedRecord := executor.Get(ctx, deletedRecordKey)
	rb.FooterSuccess("Successfully deleted DNS record %s (%s) in zone %s %s", deletedRecord.RecordName, deletedRecord.RecordID, deletedRecord.ZoneName, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
