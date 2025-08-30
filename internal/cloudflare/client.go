package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"dario.lol/cf/internal/config"
	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

var (
	Version    = "0.0.1-dev"
	ProjectURL = "https://github.com/dajooo/cloudflare-cli"
)

func UserAgent() string {
	return fmt.Sprintf("cloudflare-cli/%s (%s; %s) +%s", Version, runtime.GOOS, runtime.GOARCH, ProjectURL)
}

func NewClient() (*cloudflare.Client, error) {
	err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	if config.Cfg.APIToken != "" {
		return cloudflare.NewClient(option.WithHeader("User-Agent", UserAgent()), option.WithAPIToken(string(config.Cfg.APIToken))), nil
	}
	if config.Cfg.APIEmail != "" && config.Cfg.APIKey != "" {
		return cloudflare.NewClient(option.WithHeader("User-Agent", UserAgent()), option.WithAPIEmail(config.Cfg.APIEmail), option.WithAPIKey(string(config.Cfg.APIKey))), nil
	}
	return nil, errors.New("you need to login first. Use `cf login for that`")
}

func ZoneIDByName(client *cloudflare.Client, zoneName string) (string, error) {
	zones, err := client.Zones.List(context.Background(), zones.ZoneListParams{Name: cloudflare.String(zoneName)})
	if err != nil {
		return "", err
	}
	if len(zones.Result) == 0 {
		return "", fmt.Errorf("zone %s not found", zoneName)
	}
	return zones.Result[0].ID, nil
}

func DNSRecordIDByName(client *cloudflare.Client, zoneID string, zoneIdentifier string, recordName string) (string, error) {
	zones, err := client.DNS.Records.List(context.Background(), dns.RecordListParams{
		ZoneID: cloudflare.String(zoneID),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.String(fmt.Sprintf("%s.%s", recordName, zoneIdentifier)),
		}),
	})
	if err != nil {
		return "", err
	}
	if len(zones.Result) == 0 {
		return "", fmt.Errorf("record %s not found", recordName)
	}
	return zones.Result[0].ID, nil
}
