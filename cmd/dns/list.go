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

var recordType string
var recordName string
var recordContent string

var listCmd = &cobra.Command{
	Use:   "list <zone>",
	Short: "Lists, searches, and filters DNS records for a given zone",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, []dns.RecordResponse]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching DNS records", fetchDnsRecords).
		Display(printDnsRecords).
		Build().
		CobraRun(),
}

func init() {
	listCmd.Flags().StringVar(&recordType, "type", "", "The type of the DNS record (A, CNAME, etc.)")
	listCmd.Flags().StringVar(&recordName, "name", "", "The name of the DNS record")
	listCmd.Flags().StringVar(&recordContent, "content", "", "The content of the DNS record")
	DnsCmd.AddCommand(listCmd)
}

func fetchDnsRecords(client *cf.Client, _ *cobra.Command, args []string) ([]dns.RecordResponse, error) {
	zoneIdentifier := args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return nil, err
	}

	params := dns.RecordListParams{ZoneID: cf.F(zoneID)}

	if recordType != "" {
		params.Type = cf.F(dns.RecordListParamsType(recordType))
	}
	if recordName != "" {
		var fqdn string
		if recordName == "@" || recordName == zoneName {
			fqdn = zoneName
		} else {
			fqdn = strings.TrimSuffix(recordName, "."+zoneName) + "." + zoneName
		}
		params.Name = cf.F(dns.RecordListParamsName{Exact: cf.F(fqdn)})
	}
	if recordContent != "" {
		params.Content = cf.F(dns.RecordListParamsContent{Exact: cf.F(recordContent)})
	}

	records, err := client.DNS.Records.List(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("could not fetch DNS records: %w", err)
	}
	return records.Result, nil
}

func printDnsRecords(records []dns.RecordResponse, fetchDuration time.Duration, err error) {
	if err != nil {
		fmt.Println(ui.ErrorMessage("Failed to list DNS records", err))
		return
	}
	fmt.Println(ui.Title("DNS Records"))
	fmt.Println()

	if len(records) == 0 {
		fmt.Println(ui.Warning("No DNS records found matching your criteria"))
		return
	}

	summaryContent := fmt.Sprintf("%-12s %d", "Total:", len(records))
	fmt.Println(ui.Box(summaryContent, "Summary"))
	fmt.Println()

	for i, record := range records {
		var recordContent strings.Builder

		recordContent.WriteString(fmt.Sprintf("%-12s %s\n", "Name:", ui.Text(record.Name)))
		recordContent.WriteString(fmt.Sprintf("%-12s %s\n", "ID:", ui.Muted(record.ID)))
		recordContent.WriteString(fmt.Sprintf("%-12s %s\n", "Type:", ui.Text(string(record.Type))))
		recordContent.WriteString(fmt.Sprintf("%-12s %s", "Content:", ui.Text(record.Content)))
		if record.Proxiable {
			recordContent.WriteString("\n")
			if record.Proxied {
				recordContent.WriteString(fmt.Sprintf("%-12s %s", "Proxied", ui.Success("Yes")))
			} else {
				recordContent.WriteString(fmt.Sprintf("%-12s %s", "Proxied", ui.Error("No")))
			}
		}

		cardTitle := fmt.Sprintf("Record %d", i+1)
		if record.Name != "" {
			cardTitle = fmt.Sprintf("Record %d: %s", i+1, record.Name)
			if len(cardTitle) > 40 {
				cardTitle = fmt.Sprintf("Record %d: %s...", i+1, record.Name[:25])
			}
		}

		fmt.Println(ui.Box(recordContent.String(), cardTitle))
		fmt.Println()
	}

	fmt.Println(ui.Success(fmt.Sprintf("Found %d DNS record(s) in %v", len(records), fetchDuration)))
}
