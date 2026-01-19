package dns

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/flags"
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
		WithZone().
		WithDNSRecord().
		WithConfirmationFunc(func(ctx *executor.Context) string {
			return fmt.Sprintf("Are you sure you want to delete DNS record %s (%s) in zone %s (%s)?", executor.Get(ctx, executor.RecordNameKey), executor.Get(ctx, executor.RecordIDKey), executor.Get(ctx, executor.ZoneNameKey), executor.Get(ctx, executor.ZoneIDKey))
		}).
		Step(executor.NewStep(deletedRecordKey, "Deleting DNS record").Func(deleteDnsRecord)).
		Invalidates(func(ctx *executor.Context) []string {
			zoneID := executor.Get(ctx, executor.ZoneIDKey)
			if zoneID != "" {
				return []string{fmt.Sprintf("zone:%s", zoneID)}
			}
			return nil
		}).
		Display(printDeleteDnsResult).
		Run(),
}

func init() {
	flags.RegisterConfirmation(deleteCmd)
	DnsCmd.AddCommand(deleteCmd)
}

func deleteDnsRecord(ctx *executor.Context, _ chan<- string) (*RecordInformation, error) {
	recordID := executor.Get(ctx, executor.RecordIDKey)
	zoneID := executor.Get(ctx, executor.ZoneIDKey)
	_, err := ctx.Client.DNS.Records.Delete(context.Background(), recordID, dns.RecordDeleteParams{
		ZoneID: cf.F(zoneID),
	})
	if err != nil {
		return nil, err
	}

	return &RecordInformation{
		ZoneID:     zoneID,
		ZoneName:   executor.Get(ctx, executor.ZoneNameKey),
		RecordID:   recordID,
		RecordName: executor.Get(ctx, executor.RecordNameKey),
	}, nil
}

func printDeleteDnsResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error deleting DNS record", ctx.Error).Display()
		return
	}
	deletedRecord := executor.Get(ctx, deletedRecordKey)
	rb.FooterSuccessf("Successfully deleted DNS record %s (%s) in zone %s %s", deletedRecord.RecordName, deletedRecord.RecordID, deletedRecord.ZoneName, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
