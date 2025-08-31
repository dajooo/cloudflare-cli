package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/types"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var detailsCmd = &cobra.Command{
	Use:   "details [zone] [record]",
	Short: "Shows all details for a single DNS record",
	Args:  cobra.ExactArgs(2),
	Run: executor.NewBuilder[*cf.Client, types.DnsRecordWithZone]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching DNS record details", fetchSingleDnsRecord).
		Caches(func(cmd *cobra.Command, args []string) ([]string, error) {
			start := time.Now()
			zoneIdentifier := args[0]
			recordID := args[1]
			client, err := cloudflare.NewClient()
			if err != nil {
				return nil, err
			}
			zoneID, _, err := cloudflare.LookupZone(client, zoneIdentifier)
			if err != nil {
				return nil, err
			}
			cacheKey := fmt.Sprintf("zone:%s:record:%s", zoneID, recordID)
			fmt.Printf("Time for cache lookup %v\n", time.Since(start))
			return []string{cacheKey}, nil
		}).
		Display(printSingleDnsRecord).
		Build().
		CobraRun(),
}

func init() {
	DnsCmd.AddCommand(detailsCmd)
}

func fetchSingleDnsRecord(client *cf.Client, _ *cobra.Command, args []string, progress chan<- string) (types.DnsRecordWithZone, error) {
	zoneIdentifier := args[0]
	recordID := args[1]

	progress <- fmt.Sprintf("Looking up zone %q", zoneIdentifier)
	zoneID, zoneName, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return types.DnsRecordWithZone{}, err
	}

	progress <- fmt.Sprintf("Fetching record %q in zone %q", recordID, zoneName)
	params := dns.RecordGetParams{
		ZoneID: cf.F(zoneID),
	}
	record, err := client.DNS.Records.Get(context.Background(), recordID, params)
	if err != nil {
		return types.DnsRecordWithZone{}, fmt.Errorf("could not fetch DNS record: %w", err)
	}

	return types.DnsRecordWithZone{
		RecordResponse: *record,
		ZoneName:       zoneName,
	}, nil
}

func renderStructuredDataBox(title string, data interface{}) {
	if data == nil {
		return
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	var dataMap map[string]interface{}
	if err := json.Unmarshal(bytes, &dataMap); err != nil || len(dataMap) == 0 {
		return
	}

	icb := response.NewItemContent()
	titleCaser := cases.Title(language.English)

	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := dataMap[k]

		if v == nil || k == "JSON" {
			continue
		}

		keyFormatted := titleCaser.String(strings.ReplaceAll(k, "_", " ")) + ":"
		var valueFormatted string

		switch val := v.(type) {
		case bool:
			if val {
				valueFormatted = ui.Success("Yes")
			} else {
				valueFormatted = ui.Error("No")
			}
		case string:
			if val == "" {
				continue
			}
			valueFormatted = ui.Text(val)
		case float64:
			valueFormatted = ui.Text(strconv.FormatFloat(val, 'f', -1, 64))
		default:
			valueFormatted = ui.Text(fmt.Sprintf("%v", v))
		}
		icb.Add(keyFormatted, valueFormatted)
	}

	boxContent := icb.String()
	if boxContent != "" {
		fmt.Println()
		fmt.Println(ui.Box(boxContent, title))
	}
}

func printSingleDnsRecord(record types.DnsRecordWithZone, fetchDuration time.Duration, err error) {
	if err != nil {
		fmt.Println(ui.ErrorMessage("Failed to get DNS record", err))
		return
	}

	mainIcb := response.NewItemContent()
	mainIcb.Add("Name:", ui.Text(record.Name))
	mainIcb.Add("ID:", ui.Muted(record.ID))
	mainIcb.Add("Zone:", ui.Text(record.ZoneName))
	mainIcb.AddRaw("")
	mainIcb.Add("Type:", ui.Text(string(record.Type)))
	mainIcb.Add("Content:", ui.Text(record.Content))
	mainIcb.Add("TTL:", ui.Text(strconv.Itoa(int(record.TTL))))
	if record.Priority > 0 {
		mainIcb.Add("Priority:", ui.Text(strconv.FormatFloat(record.Priority, 'f', -1, 64)))
	}
	mainIcb.AddRaw("")
	mainIcb.Add("Proxiable:", ui.Text(fmt.Sprintf("%t", record.Proxiable)))
	if record.Proxiable {
		if record.Proxied {
			mainIcb.Add("Proxied:", ui.Success("Yes"))
		} else {
			mainIcb.Add("Proxied:", ui.Error("No"))
		}
	}
	if record.Comment != "" {
		mainIcb.AddRaw("").Add("Comment:", ui.Text(record.Comment))
	}
	if tags, ok := record.Tags.([]any); ok && len(tags) > 0 {
		var tagStrings []string
		for _, t := range tags {
			if tagStr, ok := t.(string); ok {
				tagStrings = append(tagStrings, tagStr)
			}
		}
		if len(tagStrings) > 0 {
			mainIcb.Add("Tags:", ui.Text(strings.Join(tagStrings, ", ")))
		}
	}
	mainIcb.AddRaw("")
	if !record.CreatedOn.IsZero() {
		mainIcb.Add("Created On:", ui.Muted(record.CreatedOn.Format(time.RFC1123)))
	}
	if !record.ModifiedOn.IsZero() {
		mainIcb.Add("Modified On:", ui.Muted(record.ModifiedOn.Format(time.RFC1123)))
	}
	if !record.CommentModifiedOn.IsZero() {
		mainIcb.Add("Comment Modified:", ui.Muted(record.CommentModifiedOn.Format(time.RFC1123)))
	}
	if !record.TagsModifiedOn.IsZero() {
		mainIcb.Add("Tags Modified:", ui.Muted(record.TagsModifiedOn.Format(time.RFC1123)))
	}

	fmt.Println()
	fmt.Println(ui.Box(mainIcb.String(), "DNS Record Details"))

	renderStructuredDataBox("Data", record.Data)
	renderStructuredDataBox("Settings", record.Settings)
	renderStructuredDataBox("Meta", record.Meta)

	fmt.Println()
	fmt.Println(ui.Success(fmt.Sprintf("Fetched record %s %s", record.Name, ui.Muted(fmt.Sprintf("(took %v)", fetchDuration)))))
}
