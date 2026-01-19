package ssl

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
)

var sslSetInfoKey = executor.NewKey[*SSLInfo]("sslSetInfo")

var setCmd = &cobra.Command{
	Use:   "set <zone> <mode>",
	Short: "Set SSL/TLS encryption mode (off, flexible, full, strict)",
	Args:  cobra.ExactArgs(2),
	Run: executor.New().
		WithClient().
		Step(executor.NewStep(sslSetInfoKey, "Updating SSL status").Func(setSSL)).
		Display(printSSLSetResult).
		Run(),
}

func init() {
	SslCmd.AddCommand(setCmd)
}

func setSSL(ctx *executor.Context, _ chan<- string) (*SSLInfo, error) {
	zoneIdentifier := ctx.Args[0]
	mode := ctx.Args[1]

	validModes := []string{"off", "flexible", "full", "strict"}
	if !slices.Contains(validModes, mode) {
		return nil, fmt.Errorf("invalid ssl mode: %s. valid modes are: %v", mode, validModes)
	}

	zoneID, zoneName, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %s", err)
	}

	settings, err := ctx.Client.Zones.Settings.Edit(context.Background(), "ssl", zones.SettingEditParams{
		ZoneID: cf.F(zoneID),
		Body: zones.SettingEditParamsBody{
			Value: cf.F[interface{}](mode),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error updating zone settings: %s", err)
	}

	var sslMode string
	switch v := settings.Value.(type) {
	case zones.SettingEditResponseZonesSchemasSSL:
		sslMode = string(v.Value)
	case zones.SettingEditResponseZonesSchemasSSLValue:
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

func printSSLSetResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		if strings.Contains(ctx.Error.Error(), "10000") || strings.Contains(ctx.Error.Error(), "Unauthorized") {
			rb.Error("Missing Permissions", fmt.Errorf("your API token lacks permissions. Ensure you have:\n- 'Zone / SSL and Certificates' Edit\n- 'Zone / Zone Settings' Edit")).Display()
			return
		}
		rb.Error("Error updating SSL status", ctx.Error).Display()
		return
	}
	info := executor.Get(ctx, sslSetInfoKey)
	rb.FooterSuccess("Updated SSL mode for %s to %s %s", info.ZoneName, ui.Code.Render(info.Mode), ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
