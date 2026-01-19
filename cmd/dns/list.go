package dns

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync/atomic"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/types"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	"github.com/alitto/pond/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var dnsRecordsKey = executor.NewKey[[]types.DnsRecordWithZone]("dnsRecords")

var listCmd = &cobra.Command{
	Use:   "list [zone]",
	Short: "Lists, searches, and filters DNS records for a given zone or all zones",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		allZones, _ := cmd.Flags().GetBool("all")
		if allZones && len(args) > 0 {
			return fmt.Errorf("cannot specify a zone when using the --all flag")
		}
		if !allZones && len(args) == 0 {
			return fmt.Errorf("a zone must be specified when not using the --all flag")
		}
		return nil
	},
	Run: executor.New().
		WithClient().
		WithPagination().
		WithNoCache().
		Step(executor.NewStep(dnsRecordsKey, "Fetching DNS records").Func(fetchDnsRecords)).
		Caches(func(ctx *executor.Context) ([]string, error) {
			allZones, _ := ctx.Cmd.Flags().GetBool("all")
			if allZones {
				return nil, nil
			}
			zoneID, _, err := cloudflare.LookupZone(ctx.Client, ctx.Args[0])
			if err != nil {
				return nil, err
			}
			return []string{fmt.Sprintf("zone:%s", zoneID)}, nil
		}).
		Display(printDnsRecords).
		Run(),
}

func init() {
	pagination.RegisterFlags(listCmd)
	listCmd.Flags().String("type", "", "The type of the DNS record (A, CNAME, etc.)")
	listCmd.Flags().String("name", "", "The name of the DNS record")
	listCmd.Flags().String("content", "", "The content of the DNS record")
	listCmd.Flags().BoolP("all", "A", false, "List records across all zones")
	listCmd.Flags().BoolP("compact", "c", false, "Display output in a compact table format")
	listCmd.Flags().Bool("no-cache", false, "Don't use the cache when listing records")
	DnsCmd.AddCommand(listCmd)
}

func fetchDnsRecords(ctx *executor.Context, progress chan<- string) ([]types.DnsRecordWithZone, error) {
	allZones, _ := ctx.Cmd.Flags().GetBool("all")
	recordType, _ := ctx.Cmd.Flags().GetString("type")
	recordName, _ := ctx.Cmd.Flags().GetString("name")
	recordContent, _ := ctx.Cmd.Flags().GetString("content")

	if allZones {
		progress <- "Fetching list of all zones"
		var zoneList []zones.Zone
		pager := ctx.Client.Zones.ListAutoPaging(context.Background(), zones.ZoneListParams{})
		for pager.Next() {
			zoneList = append(zoneList, pager.Current())
		}
		if err := pager.Err(); err != nil {
			return nil, fmt.Errorf("could not list zones: %w", err)
		}

		pool := pond.NewResultPool[[]types.DnsRecordWithZone](10)
		group := pool.NewGroup()

		totalZones := len(zoneList)
		var completed atomic.Int32

		for _, zone := range zoneList {
			zone := zone
			group.SubmitErr(func() ([]types.DnsRecordWithZone, error) {
				records, err := getRecordsForZone(ctx.Client, zone.ID, zone.Name, recordType, recordName, recordContent)
				if err != nil {
					return nil, err
				}
				recordsWithZone := make([]types.DnsRecordWithZone, len(records))
				for i, r := range records {
					recordsWithZone[i] = types.DnsRecordWithZone{
						RecordResponse: r,
						ZoneName:       zone.Name,
					}
				}
				c := completed.Add(1)
				progress <- fmt.Sprintf("Fetched records for %d/%d zones", c, totalZones)
				return recordsWithZone, err
			})
		}

		nestedResults, err := group.Wait()
		if err != nil {
			return nil, err
		}

		var allRecords []types.DnsRecordWithZone
		for _, resultSlice := range nestedResults {
			allRecords = append(allRecords, resultSlice...)
		}

		return allRecords, nil
	}

	zoneIdentifier := ctx.Args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
	if err != nil {
		return nil, err
	}
	records, err := getRecordsForZone(ctx.Client, zoneID, zoneName, recordType, recordName, recordContent)
	if err != nil {
		return nil, err
	}
	recordsWithZone := make([]types.DnsRecordWithZone, len(records))
	for i, r := range records {
		recordsWithZone[i] = types.DnsRecordWithZone{
			RecordResponse: r,
			ZoneName:       zoneName,
		}
	}
	return recordsWithZone, nil
}

func getRecordsForZone(client *cf.Client, zoneID, zoneName, recordType, recordName, recordContent string) ([]dns.RecordResponse, error) {
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
		return nil, fmt.Errorf("could not fetch DNS records for zone %s: %w", zoneName, err)
	}
	return records.Result, nil
}

func printDnsRecords(ctx *executor.Context) {
	if ctx.Error != nil {
		fmt.Println(ui.ErrorMessage("Failed to list DNS records", ctx.Error))
		return
	}

	records := executor.Get(ctx, dnsRecordsKey)
	paginated, info := pagination.Paginate(records, ctx.Pagination)

	if len(paginated) == 0 {
		fmt.Println(ui.Warning("No DNS records found matching your criteria"))
		return
	}

	compact, _ := ctx.Cmd.Flags().GetBool("compact")
	allZones, _ := ctx.Cmd.Flags().GetBool("all")

	if compact {
		printDnsRecordsCompact(paginated)
	} else {
		printDnsRecordsCards(paginated, allZones)
	}

	fmt.Println()
	footer := fmt.Sprintf("Showing %d of %d DNS record(s)", info.Showing, info.Total)
	footer += " " + ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))
	fmt.Println(ui.Success(footer))
}

func printDnsRecordsCompact(records []types.DnsRecordWithZone) {
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 100
	}

	headers := []string{"ZONE", "NAME", "TYPE", "CONTENT", "PROXIED", "ID"}
	var data [][]string
	for _, r := range records {
		proxiedText := ui.BodySmall.Render("No")
		if r.Proxied {
			proxiedText = ui.StatusSuccess.Render("Yes")
		}
		if !r.Proxiable {
			proxiedText = ui.BodySmall.Render("N/A")
		}
		data = append(data, []string{
			r.ZoneName,
			r.Name,
			string(r.Type),
			r.Content,
			proxiedText,
			r.ID,
		})
	}

	t := table.New().
		Headers(headers...).
		Rows(data...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ui.C.Gray500)).
		Width(termWidth - 2)

	t.StyleFunc(func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)

		switch col {
		case 1:
			style = style.MaxWidth(25)
		case 2:
			style = style.Width(8)
		case 3:
			style = style.MaxWidth(30)
		case 4:
			style = style.Width(10)
		case 5:
			style = style.MaxWidth(32).Foreground(ui.C.Gray400)
		}

		if row == -1 {
			return style.Foreground(ui.C.Primary500).Bold(true)
		}

		return style
	})

	fmt.Println(t)
}

func printDnsRecordsCards(records []types.DnsRecordWithZone, allZones bool) {
	fmt.Println(ui.Title("DNS Records"))
	fmt.Println()

	summaryContent := fmt.Sprintf("%-12s %d", "Total:", len(records))
	fmt.Println(ui.Box(summaryContent, "Summary"))
	fmt.Println()

	if allZones {
		recordsByZone := make(map[string][]types.DnsRecordWithZone)
		for _, record := range records {
			recordsByZone[record.ZoneName] = append(recordsByZone[record.ZoneName], record)
		}

		zoneNames := make([]string, 0, len(recordsByZone))
		for name := range recordsByZone {
			zoneNames = append(zoneNames, name)
		}
		sort.Strings(zoneNames)

		for i, zoneName := range zoneNames {
			if i > 0 {
				fmt.Println()
			}
			fmt.Println(ui.DividerWithText(zoneName, 80))
			fmt.Println()
			for _, record := range recordsByZone[zoneName] {
				fmt.Println(renderRecordCard(record))
			}
		}
	} else {
		for _, record := range records {
			fmt.Println(renderRecordCard(record))
		}
	}
}

func renderRecordCard(record types.DnsRecordWithZone) string {
	icb := response.NewItemContent().
		Add("Name:", ui.Text(record.Name)).
		Add("ID:", ui.Muted(record.ID)).
		Add("Type:", ui.Text(string(record.Type))).
		Add("Content:", ui.Text(record.Content))

	if record.Proxiable {
		icb.AddRaw("")
		if record.Proxied {
			icb.Add("Proxied", ui.Success("Yes"))
		} else {
			icb.Add("Proxied", ui.Error("No"))
		}
	}

	return ui.Box(icb.String(), record.Name)
}
