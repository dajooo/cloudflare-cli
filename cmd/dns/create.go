package dns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
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
		Fetch("Fetching DNS records", createDnsRecord).
		Display(func(record *RecordInformation, duration time.Duration, err error) {
			if err != nil {
				fmt.Println(ui.ErrorMessage("Error deleting DNS record", err))
				return
			}
			fmt.Println(ui.Success(fmt.Sprintf("Successfully created DNS record %s.%s (%s)", record.RecordName, record.ZoneName, record.RecordID)))
		}).
		Build().
		CobraRun(),
}

func createDnsRecord(client *cf.Client, _ *cobra.Command, args []string) (*RecordInformation, error) {
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

func init() {
	createCmd.Flags().IntVar(&ttl, "ttl", 1, "The TTL of the DNS record")
	createCmd.Flags().BoolVar(&proxied, "proxied", false, "Whether the DNS record should be proxied")
	DnsCmd.AddCommand(createCmd)
}
