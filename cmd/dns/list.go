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

var recordType string
var recordName string
var recordContent string

var listCmd = &cobra.Command{
	Use:   "list <zone>",
	Short: "Lists, searches, and filters DNS records for a given zone",
	Args:  cobra.ExactArgs(1),
	Run:   executeDnsList,
}

func init() {
	listCmd.Flags().StringVar(&recordType, "type", "", "The type of the DNS record (A, CNAME, etc.)")
	listCmd.Flags().StringVar(&recordName, "name", "", "The name of the DNS record")
	listCmd.Flags().StringVar(&recordContent, "content", "", "The content of the DNS record")
	DnsCmd.AddCommand(listCmd)
}

func executeDnsList(cmd *cobra.Command, args []string) {
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

	params := dns.RecordListParams{
		ZoneID: cf.F(zoneID),
	}
	if recordType != "" {
		params.Type = cf.F(dns.RecordListParamsType(recordType))
	}
	if recordName != "" {
		params.Name = cf.F(dns.RecordListParamsName{Exact: cf.F(recordName)})
	}
	if recordContent != "" {
		params.Content = cf.F(dns.RecordListParamsContent{Exact: cf.F(recordContent)})
	}

	records, err := client.DNS.Records.List(context.Background(), params)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error fetching DNS records", err))
		return
	}

	printDnsRecords(records.Result)
}

func printDnsRecords(records []dns.RecordResponse) {
	fmt.Println(ui.Title("DNS Records"))
	fmt.Println()

	if len(records) == 0 {
		fmt.Println(ui.Warning("No DNS records found"))
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

	fmt.Println(ui.Success(fmt.Sprintf("Found %d DNS record(s)", len(records))))
}
