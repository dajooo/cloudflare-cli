package zone

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
)

var accountID string
var status string
var noCache bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones in an account",
	Run: executor.NewBuilder[*cf.Client, []zones.Zone]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching zones", fetchZones).
		SkipCache(noCache).
		Caches(func(cmd *cobra.Command, args []string) ([]string, error) {
			return []string{"zones:list"}, nil
		}).
		Display(printZonesList).
		Build().
		CobraRun(),
}

func init() {
	listCmd.Flags().StringVar(&status, "status", "", "The status of the zones to list (active, pending)")
	listCmd.Flags().BoolVar(&noCache, "no-cache", false, "Don't use the cache when listing records")
	ZoneCmd.AddCommand(listCmd)
}

func fetchZones(client *cf.Client, cmd *cobra.Command, _ []string, _ chan<- string) ([]zones.Zone, error) {
	params := zones.ZoneListParams{}

	accID, err := cloudflare.GetAccountID(client, cmd, accountID)
	// For listing, if GetAccountID returns error (e.g. multiple found), we might default to empty?
	// But enforcing context is better.
	if err != nil {
		return nil, err
	}
	params.Account.Value.ID.Value = accID

	if status != "" {
		params.Status.Value = zones.ZoneListParamsStatus(status)
	}

	zonesList, err := client.Zones.List(context.Background(), params)
	if err != nil {
		return nil, err
	}

	for _, zone := range zonesList.Result {
		cloudflare.SetID(cloudflare.ZoneCacheKey(zone.Name), zone.ID)
		cloudflare.SetID(cloudflare.ZoneCacheKey(zone.ID), zone.Name)
	}

	return zonesList.Result, nil
}

func printZonesList(zonesList []zones.Zone, fetchDuration time.Duration, err error) {
	rb := response.New().
		Title("Accessible Zones").
		NoItemsMessage("No zones found")

	if err != nil {
		rb.Error("Error fetching zones", err).Display()
		return
	}

	rb.Summary("Total:", len(zonesList))

	for i, zone := range zonesList {
		icb := response.NewItemContent().
			Add("Name:", ui.Text(zone.Name)).
			Add("ID:", ui.Muted(zone.ID)).
			Add("Status:", ui.Text(string(zone.Status)))

		if !zone.CreatedOn.IsZero() {
			icb.Add("Created:", ui.Small(zone.CreatedOn.Format("2006-01-02 15:04:05")))
		}

		cardTitle := fmt.Sprintf("Zone %d: %s", i+1, zone.Name)
		rb.AddItem(cardTitle, icb.String())
	}

	if len(zonesList) > 0 {
		rb.FooterSuccess("Found %d accessible zone(s) %s", len(zonesList), ui.Muted(fmt.Sprintf("(took %v)", fetchDuration)))
	}

	rb.Display()
}
