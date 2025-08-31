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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones in an account",
	Run: executor.NewBuilder[*cf.Client, []zones.Zone]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching zones", fetchZones).
		Display(printZonesList).
		Build().
		CobraRun(),
}

func init() {
	listCmd.Flags().StringVar(&accountID, "account-id", "", "The account ID to list zones for")
	listCmd.Flags().StringVar(&status, "status", "", "The status of the zones to list (active, pending)")
	ZoneCmd.AddCommand(listCmd)
}

func fetchZones(client *cf.Client, _ *cobra.Command, _ []string, _ chan<- string) ([]zones.Zone, error) {
	params := zones.ZoneListParams{}
	if accountID != "" {
		params.Account.Value.ID.Value = accountID
	}
	if status != "" {
		params.Status.Value = zones.ZoneListParamsStatus(status)
	}

	zonesList, err := client.Zones.List(context.Background(), params)
	if err != nil {
		return nil, err
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
		rb.FooterSuccess("Found %d accessible zone(s) in %v", len(zonesList), fetchDuration)
	}

	rb.Display()
}
