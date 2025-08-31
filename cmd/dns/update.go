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

var updateRecordName string
var updateRecordType string
var updateRecordContent string
var updateProxied bool

type updateResult struct {
	Record *dns.RecordResponse
	ZoneID string
}

var updateCmd = &cobra.Command{
	Use:   "update <zone> <record_id>",
	Short: "Updates an existing DNS record, identified by its ID",
	Args:  cobra.ExactArgs(2),
	Run: executor.NewBuilder[*cf.Client, *updateResult]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Updating DNS record", updateDnsRecord).
		Display(printUpdateDnsResult).
		Invalidates(func(cmd *cobra.Command, args []string, result *updateResult) []string {
			return []string{fmt.Sprintf("zone:%s", result.ZoneID)}
		}).
		Build().
		CobraRun(),
}

func init() {
	updateCmd.Flags().StringVar(&updateRecordName, "name", "", "The new name of the DNS record")
	updateCmd.Flags().StringVar(&updateRecordType, "type", "", "The new type of the DNS record")
	updateCmd.Flags().StringVar(&updateRecordContent, "content", "", "The new content of the DNS record")
	updateCmd.Flags().BoolVar(&updateProxied, "proxied", false, "Whether the DNS record should be proxied")
	DnsCmd.AddCommand(updateCmd)
}

func updateDnsRecord(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*updateResult, error) {
	zoneIdentifier := args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %w", err)
	}

	recordIdentifier := args[1]
	recordID, err := cloudflare.LookupDNSRecordID(client, zoneID, zoneName, recordIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding record: %w", err)
	}

	if updateRecordType == "" || updateRecordName == "" || updateRecordContent == "" {
		return nil, fmt.Errorf("flags --name, --type, and --content are required for an update")
	}

	var body dns.RecordUpdateParamsBodyUnion
	switch strings.ToUpper(updateRecordType) {
	case "A":
		body = &dns.ARecordParam{
			Type:    cf.F(dns.ARecordTypeA),
			Name:    cf.F(updateRecordName),
			Content: cf.F(strings.ReplaceAll(updateRecordContent, "@", zoneName)),
			Proxied: cf.F(updateProxied),
		}
	case "AAAA":
		body = &dns.AAAARecordParam{
			Type:    cf.F(dns.AAAARecordTypeAAAA),
			Name:    cf.F(updateRecordName),
			Content: cf.F(strings.ReplaceAll(updateRecordContent, "@", zoneName)),
			Proxied: cf.F(updateProxied),
		}
	case "CNAME":
		body = &dns.CNAMERecordParam{
			Type:    cf.F(dns.CNAMERecordTypeCNAME),
			Name:    cf.F(updateRecordName),
			Content: cf.F(strings.ReplaceAll(updateRecordContent, "@", zoneName)),
			Proxied: cf.F(updateProxied),
		}
	default:
		return nil, fmt.Errorf("unsupported record type: %s", updateRecordType)
	}

	params := dns.RecordUpdateParams{
		ZoneID: cf.F(zoneID),
		Body:   body,
	}

	record, err := client.DNS.Records.Update(context.Background(), recordID, params)
	if err != nil {
		return nil, err
	}

	return &updateResult{
		Record: record,
		ZoneID: zoneID,
	}, nil
}

func printUpdateDnsResult(result *updateResult, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error updating DNS record", err).Display()
		return
	}
	rb.FooterSuccess("Successfully updated DNS record %s (%s) %s", result.Record.Name, result.Record.ID, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
