package ssl

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
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <zone>",
	Short: "Get SSL/TLS encryption mode",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, *SSLInfo]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching SSL status", getSSL).
		Display(printSSLResult).
		Build().
		CobraRun(),
}

type SSLInfo struct {
	ZoneID   string
	ZoneName string
	Mode     string
}

func init() {
	SslCmd.AddCommand(getCmd)
}

func getSSL(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*SSLInfo, error) {
	zoneIdentifier := args[0]
	zoneID, zoneName, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %s", err)
	}

	// Fetch zone settings
	settings, err := client.Zones.Settings.Get(context.Background(), "ssl", zones.SettingGetParams{
		ZoneID: cf.F(zoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching zone settings: %s", err)
	}

	// Handle different response types
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

func printSSLResult(info *SSLInfo, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		if strings.Contains(err.Error(), "10000") || strings.Contains(err.Error(), "Unauthorized") {
			rb.Error("Missing Permissions", fmt.Errorf("your API token lacks permissions. Ensure you have:\n- 'Zone / SSL and Certificates' Edit\n- 'Zone / Zone Settings' Read")).Display()
			return
		}
		rb.Error("Error fetching SSL status", err).Display()
		return
	}
	rb.FooterSuccess("SSL mode for %s is %s %s", info.ZoneName, ui.Code.Render(info.Mode), ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
