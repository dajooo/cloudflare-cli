package dns

import (
	"context"
	"fmt"
	"strings"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/ui"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/spf13/cobra"
)

var updateRecordName string
var updateRecordType string
var updateRecordContent string
var updateProxied bool

var updateCmd = &cobra.Command{
	Use:   "update <zone> <record_id>",
	Short: "Updates an existing DNS record, identified by its ID",
	Args:  cobra.ExactArgs(2),
	Run:   executeDnsUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateRecordName, "name", "", "The new name of the DNS record")
	updateCmd.Flags().StringVar(&updateRecordType, "type", "", "The new type of the DNS record")
	updateCmd.Flags().StringVar(&updateRecordContent, "content", "", "The new content of the DNS record")
	updateCmd.Flags().BoolVar(&updateProxied, "proxied", false, "Whether the DNS record should be proxied")
	DnsCmd.AddCommand(updateCmd)
}

func executeDnsUpdate(cmd *cobra.Command, args []string) {
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

	var body dns.RecordUpdateParamsBodyUnion
	switch strings.ToUpper(updateRecordType) {
	case "A":
		body = &dns.ARecordParam{
			Type:    cf.F(dns.ARecordTypeA),
			Name:    cf.F(updateRecordName),
			Content: cf.F(strings.ReplaceAll(updateRecordContent, "@", zoneIdentifier)),
			Proxied: cf.F(updateProxied),
		}
	case "AAAA":
		body = &dns.AAAARecordParam{
			Type:    cf.F(dns.AAAARecordTypeAAAA),
			Name:    cf.F(updateRecordName),
			Content: cf.F(strings.ReplaceAll(updateRecordContent, "@", zoneIdentifier)),
			Proxied: cf.F(updateProxied),
		}
	case "CNAME":
		body = &dns.CNAMERecordParam{
			Type:    cf.F(dns.CNAMERecordTypeCNAME),
			Name:    cf.F(updateRecordName),
			Content: cf.F(strings.ReplaceAll(updateRecordContent, "@", zoneIdentifier)),
			Proxied: cf.F(updateProxied),
		}
	default:
		fmt.Println(ui.ErrorMessage(fmt.Sprintf("Unsupported record type: %s", updateRecordType)))
		return
	}

	params := dns.RecordUpdateParams{
		ZoneID: cf.F(zoneID),
		Body:   body,
	}

	record, err := client.DNS.Records.Update(context.Background(), recordID, params)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error updating DNS record", err))
		return
	}

	fmt.Println(ui.Success(fmt.Sprintf("Successfully updated DNS record %s", record.Name)))
}
