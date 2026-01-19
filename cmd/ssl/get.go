package ssl

import (
	"context"
	"fmt"
	"strings"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
)

var sslInfoKey = executor.NewKey[*SSLInfo]("sslInfo")

var getCmd = &cobra.Command{
	Use:   "get <zone>",
	Short: "Get SSL/TLS encryption mode",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		Step(executor.NewStep(sslInfoKey, "Fetching SSL status").
			Func(getSSL).
			CacheKeyFunc(func(ctx *executor.Context) string {
				zoneID, _, err := cloudflare.LookupZone(ctx.Client, ctx.Args[0])
				if err != nil {
					return ""
				}
				return fmt.Sprintf("zone:%s:ssl", zoneID)
			})).
		Display(printSSLResult).
		Run(),
}

type SSLInfo struct {
	ZoneID   string
	ZoneName string
	Mode     string
}

func init() {
	SslCmd.AddCommand(getCmd)
}

func getSSL(ctx *executor.Context, _ chan<- string) (*SSLInfo, error) {
	zoneIdentifier := ctx.Args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %s", err)
	}

	settings, err := ctx.Client.Zones.Settings.Get(context.Background(), "ssl", zones.SettingGetParams{
		ZoneID: cf.F(zoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching zone settings: %s", err)
	}

	var sslMode string
	switch v := settings.Value.(type) {
	case zones.SettingGetResponseZonesSchemasSSL:
		sslMode = string(v.Value)
	case zones.SettingGetResponseZonesSchemasSSLValue:
		sslMode = string(v)
	default:
		return nil, fmt.Errorf("unexpected response type for SSL setting: %T", settings.Value)
	}

	return &SSLInfo{
		ZoneID:   zoneID,
		ZoneName: zoneName,
		Mode:     sslMode,
	}, nil
}

func printSSLResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		if strings.Contains(ctx.Error.Error(), "10000") || strings.Contains(ctx.Error.Error(), "Unauthorized") {
			rb.Error("Missing Permissions", fmt.Errorf("your API token lacks permissions. Ensure you have:\n- 'Zone / SSL and Certificates' Edit\n- 'Zone / Zone Settings' Read")).Display()
			return
		}
		rb.Error("Error fetching SSL status", ctx.Error).Display()
		return
	}
	info := executor.Get(ctx, sslInfoKey)
	rb.FooterSuccessf("SSL mode for %s is %s %s", info.ZoneName, ui.Code.Render(info.Mode), ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
