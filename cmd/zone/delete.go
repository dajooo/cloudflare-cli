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

var yes bool

type DeletedZoneInfo struct {
	ID   string
	Name string
}

var deleteCmd = &cobra.Command{
	Use:   "delete <zone_name|zone_id>",
	Short: "Deletes a zone",
	Args:  cobra.ExactArgs(1),
	Run:   executeZoneDelete,
}

func init() {
	deleteCmd.Flags().BoolVar(&yes, "yes", false, "Bypass the confirmation prompt")
	ZoneCmd.AddCommand(deleteCmd)
}

func executeZoneDelete(cmd *cobra.Command, args []string) {
	zoneIdentifier := args[0]

	if !yes {
		confirmed, err := ui.Confirm(fmt.Sprintf("Are you sure you want to delete zone %s?", zoneIdentifier))
		if err != nil || !confirmed {
			fmt.Println(ui.Warning("Zone deletion cancelled."))
			return
		}
	}

	executor.NewBuilder[*cf.Client, *DeletedZoneInfo]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Deleting zone", deleteZone).
		Display(printDeleteZoneResult).
		Build().
		CobraRun()(cmd, args)
}

func deleteZone(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*DeletedZoneInfo, error) {
	zoneIdentifier := args[0]

	zoneID, zoneName, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %w", err)
	}

	_, err = client.Zones.Delete(context.Background(), zones.ZoneDeleteParams{ZoneID: cf.F(zoneID)})
	if err != nil {
		return nil, err
	}

	return &DeletedZoneInfo{
		ID:   zoneID,
		Name: zoneName,
	}, nil
}

func printDeleteZoneResult(zone *DeletedZoneInfo, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error deleting zone", err).Display()
		return
	}
	rb.FooterSuccess("Successfully deleted zone %s (%s) %s", zone.Name, zone.ID, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
