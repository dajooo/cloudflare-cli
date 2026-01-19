package dns

import (
	"context"
	"fmt"
	"strings"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/spf13/cobra"
)

var createdRecordKey = executor.NewKey[*RecordInformation]("createdRecord")

var createCmd = &cobra.Command{
	Use:   "create <zone> <name> <type> <content>",
	Short: "Creates a new DNS record",
	Args:  cobra.ExactArgs(4),
	Run: executor.New().
		WithClient().
		Step(executor.NewStep(createdRecordKey, "Creating DNS record").Func(createDnsRecord)).
		Invalidates(func(ctx *executor.Context) []string {
			record := executor.Get(ctx, createdRecordKey)
			if record != nil {
				return []string{fmt.Sprintf("zone:%s", record.ZoneID)}
			}
			return nil
		}).
		Display(printCreateDnsResult).
		Run(),
}

func init() {
	createCmd.Flags().Int("ttl", 1, "The TTL of the DNS record")
	createCmd.Flags().Bool("proxied", false, "Whether the DNS record should be proxied")
	DnsCmd.AddCommand(createCmd)
}

func createDnsRecord(ctx *executor.Context, _ chan<- string) (*RecordInformation, error) {
	zoneIdentifier := ctx.Args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %s", err)
	}

	recordName := ctx.Args[1]
	recordType := ctx.Args[2]
	recordContent := strings.ReplaceAll(ctx.Args[3], "@", zoneName)
	ttl, _ := ctx.Cmd.Flags().GetInt("ttl")
	proxied, _ := ctx.Cmd.Flags().GetBool("proxied")

	var body dns.RecordNewParamsBodyUnion
	switch strings.ToUpper(recordType) {
	case "A":
		body = &dns.ARecordParam{
			Type:    cf.F(dns.ARecordTypeA),
			Name:    cf.F(recordName),
			Content: cf.F(recordContent),
			TTL:     cf.F(dns.TTL(ttl)),
			Proxied: cf.F(proxied),
		}
	case "AAAA":
		body = &dns.AAAARecordParam{
			Type:    cf.F(dns.AAAARecordTypeAAAA),
			Name:    cf.F(recordName),
			Content: cf.F(recordContent),
			TTL:     cf.F(dns.TTL(ttl)),
			Proxied: cf.F(proxied),
		}
	case "CNAME":
		body = &dns.CNAMERecordParam{
			Type:    cf.F(dns.CNAMERecordTypeCNAME),
			Name:    cf.F(recordName),
			Content: cf.F(recordContent),
			TTL:     cf.F(dns.TTL(ttl)),
			Proxied: cf.F(proxied),
		}
	default:
		return nil, fmt.Errorf("unsupported record type: %s", recordType)
	}

	params := dns.RecordNewParams{
		ZoneID: cf.F(zoneID),
		Body:   body,
	}

	record, err := ctx.Client.DNS.Records.New(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("error creating DNS record: %s", err)
	}

	return &RecordInformation{
		ZoneID:     zoneID,
		ZoneName:   zoneName,
		RecordID:   record.ID,
		RecordName: record.Name,
	}, nil
}

func printCreateDnsResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error creating DNS record", ctx.Error).Display()
		return
	}
	record := executor.Get(ctx, createdRecordKey)
	rb.FooterSuccessf("Successfully created DNS record %s (%s) in zone %s %s", record.RecordName, record.RecordID, record.ZoneName, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
