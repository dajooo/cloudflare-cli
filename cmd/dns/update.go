package dns

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/spf13/cobra"
)

type updateResult struct {
	Record *dns.RecordResponse
	ZoneID string
}

var updatedRecordKey = executor.NewKey[*updateResult]("updatedRecord")

var updateCmd = &cobra.Command{
	Use:   "update <zone> <record>",
	Short: "Updates an existing DNS record, identified by its ID",
	Args:  cobra.ExactArgs(2),
	Run: executor.New().
		WithClient().
		Step(executor.NewStep(updatedRecordKey, "Updating DNS record").Func(updateDnsRecord)).
		Invalidates(func(ctx *executor.Context) []string {
			result := executor.Get(ctx, updatedRecordKey)
			if result != nil {
				return []string{fmt.Sprintf("zone:%s:", result.ZoneID)}
			}
			return nil
		}).
		Display(printUpdateDnsResult).
		Run(),
}

func init() {
	updateCmd.Flags().String("name", "", "The new name of the DNS record")
	updateCmd.Flags().String("type", "", "The new type of the DNS record")
	updateCmd.Flags().String("content", "", "The new content of the DNS record")
	updateCmd.Flags().Bool("proxied", false, "Whether the DNS record should be proxied")
	DnsCmd.AddCommand(updateCmd)
}

func updateDnsRecord(ctx *executor.Context, _ chan<- string) (*updateResult, error) {
	zoneIdentifier := ctx.Args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %w", err)
	}

	recordIdentifier := ctx.Args[1]
	recordID, err := cloudflare.LookupDNSRecordID(ctx.Client, zoneID, zoneName, recordIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding record: %w", err)
	}

	updateRecordName, _ := ctx.Cmd.Flags().GetString("name")
	updateRecordType, _ := ctx.Cmd.Flags().GetString("type")
	updateRecordContent, _ := ctx.Cmd.Flags().GetString("content")
	updateProxied, _ := ctx.Cmd.Flags().GetBool("proxied")

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
	case "TXT":
		body = &dns.TXTRecordParam{
			Type:    cf.F(dns.TXTRecordTypeTXT),
			Name:    cf.F(updateRecordName),
			Content: cf.F(updateRecordContent),
		}
	case "MX":
		parts := strings.Fields(updateRecordContent)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid MX record content. Expected: '<priority> <target>'")
		}
		priority, err := strconv.ParseUint(parts[0], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid MX priority: %w", err)
		}
		body = &dns.MXRecordParam{
			Type:     cf.F(dns.MXRecordTypeMX),
			Name:     cf.F(updateRecordName),
			Content:  cf.F(strings.ReplaceAll(strings.Join(parts[1:], " "), "@", zoneName)),
			Priority: cf.F(float64(priority)),
		}
	case "SRV":
		parts := strings.Fields(updateRecordContent)
		var p, w, po uint64
		var target string
		var err error

		if len(parts) == 3 {
			p, err = strconv.ParseUint(parts[0], 10, 16)
			if err == nil {
				w = 0
				po, err = strconv.ParseUint(parts[1], 10, 16)
				target = strings.Join(parts[2:], " ")
			}
		} else if len(parts) >= 4 {
			p, err = strconv.ParseUint(parts[0], 10, 16)
			if err == nil {
				w, err = strconv.ParseUint(parts[1], 10, 16)
				if err == nil {
					po, err = strconv.ParseUint(parts[2], 10, 16)
					target = strings.Join(parts[3:], " ")
				}
			}
		} else {
			return nil, fmt.Errorf("invalid SRV record content. Expected: '<priority> <weight> <port> <target>'")
		}

		if err != nil {
			return nil, fmt.Errorf("error parsing SRV record fields: %w", err)
		}

		body = &dns.SRVRecordParam{
			Type: cf.F(dns.SRVRecordTypeSRV),
			Name: cf.F(updateRecordName),
			Data: cf.F(dns.SRVRecordDataParam{
				Priority: cf.F(float64(p)),
				Weight:   cf.F(float64(w)),
				Port:     cf.F(float64(po)),
				Target:   cf.F(strings.ReplaceAll(target, "@", zoneName)),
			}),
		}
	default:
		return nil, fmt.Errorf("unsupported record type: %s", updateRecordType)
	}

	params := dns.RecordUpdateParams{
		ZoneID: cf.F(zoneID),
		Body:   body,
	}

	record, err := ctx.Client.DNS.Records.Update(context.Background(), recordID, params)
	if err != nil {
		return nil, err
	}

	return &updateResult{
		Record: record,
		ZoneID: zoneID,
	}, nil
}

func printUpdateDnsResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error updating DNS record", ctx.Error).Display()
		return
	}
	result := executor.Get(ctx, updatedRecordKey)
	rb.FooterSuccessf("Successfully updated DNS record %s (%s) %s", result.Record.Name, result.Record.ID, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
