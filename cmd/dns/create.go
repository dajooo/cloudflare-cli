package dns

import (
	"context"
	"fmt"
	"strings"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/dajooo/cloudflare-cli/internal/cloudflare"
	"github.com/dajooo/cloudflare-cli/internal/ui"
	"github.com/spf13/cobra"
)

var ttl int
var proxied bool

var createCmd = &cobra.Command{
	Use:   "create <zone> <name> <type> <content>",
	Short: "Creates a new DNS record",
	Args:  cobra.ExactArgs(4),
	Run:   executeDnsCreate,
}

func init() {
	createCmd.Flags().IntVar(&ttl, "ttl", 1, "The TTL of the DNS record")
	createCmd.Flags().BoolVar(&proxied, "proxied", false, "Whether the DNS record should be proxied")
	DnsCmd.AddCommand(createCmd)
}

func executeDnsCreate(cmd *cobra.Command, args []string) {
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
	recordType := args[2]
	recordContent := strings.ReplaceAll(args[3], "@", zoneIdentifier)

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
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Unsupported record type: %s", recordType)))
		return
	}

	params := dns.RecordNewParams{
		ZoneID: cf.F(zoneID),
		Body:   body,
	}

	record, err := client.DNS.Records.New(context.Background(), params)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error creating DNS record", err))
		return
	}

	fmt.Println(ui.Success(fmt.Sprintf("Successfully created DNS record %s", record.Name)))
}
