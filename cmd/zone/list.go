package zone

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
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

func fetchZones(client *cf.Client, _ *cobra.Command, _ []string) ([]zones.Zone, error) {
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
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error fetching zones", err))
		return
	}

	fmt.Println(ui.Title("Accessible Zones"))
	fmt.Println()

	if len(zonesList) == 0 {
		fmt.Println(ui.Warning("No zones found"))
		return
	}

	summaryContent := fmt.Sprintf("%-12s %d", "Total:", len(zonesList))
	fmt.Println(ui.Box(summaryContent, "Summary"))
	fmt.Println()

	for i, zone := range zonesList {
		var zoneContent strings.Builder

		zoneContent.WriteString(fmt.Sprintf("%-12s %s\n", "Name:", ui.Text(zone.Name)))
		zoneContent.WriteString(fmt.Sprintf("%-12s %s\n", "ID:", ui.Muted(zone.ID)))
		zoneContent.WriteString(fmt.Sprintf("%-12s %s", "Status:", ui.Text(string(zone.Status))))

		if !zone.CreatedOn.IsZero() {
			zoneContent.WriteString(fmt.Sprintf("\n%-12s %s", "Created:", ui.Small(zone.CreatedOn.Format("2006-01-02 15:04:05"))))
		}

		cardTitle := fmt.Sprintf("Zone %d", i+1)
		if zone.Name != "" {
			cardTitle = fmt.Sprintf("Zone %d: %s", i+1, zone.Name)
			if len(cardTitle) > 40 {
				cardTitle = fmt.Sprintf("Zone %d: %s...", i+1, zone.Name[:25])
			}
		}

		fmt.Println(ui.Box(zoneContent.String(), cardTitle))
		fmt.Println()
	}

	fmt.Println(ui.Success(fmt.Sprintf("Found %d accessible zone(s) in %v", len(zonesList), fetchDuration)))
}
