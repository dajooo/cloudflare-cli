package dns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/spf13/cobra"
)

var ttl int
var proxied bool

var createCmd = &cobra.Command{
	Use:   "create <zone> <name> <type> <content>",
	Short: "Creates a new DNS record",
	Args:  cobra.ExactArgs(4),
	Run: executor.NewBuilder[*cf.Client, *RecordInformation]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Creating DNS record", createDnsRecord).
		Display(printCreateDnsResult).
		Invalidates(func(cmd *cobra.Command, args []string, result *RecordInformation) []string {
			return []string{fmt.Sprintf("zone:%s", result.ZoneID)}
		}).
		Build().
		CobraRun(),
}

func init() {
	createCmd.Flags().IntVar(&ttl, "ttl", 1, "The TTL of the DNS record")
	createCmd.Flags().BoolVar(&proxied, "proxied", false, "Whether the DNS record should be proxied")
	DnsCmd.AddCommand(createCmd)
}

func createDnsRecord(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*RecordInformation, error) {
	zoneIdentifier := args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %s", err)
	}

	recordName := args[1]
	recordType := args[2]
	recordContent := strings.ReplaceAll(args[3], "@", zoneName)

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

	record, err := client.DNS.Records.New(context.Background(), params)
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

func printCreateDnsResult(record *RecordInformation, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error creating DNS record", err).Display()
		return
	}
	rb.FooterSuccess("Successfully created DNS record %s (%s) in zone %s %s", record.RecordName, record.RecordID, record.ZoneName, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
