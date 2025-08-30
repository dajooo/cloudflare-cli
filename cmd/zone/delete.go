package zone

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/ui"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
)

var yes bool

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
	client, err := cloudflare.NewClient()
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error loading config", err))
		return
	}

	zoneIdentifier := args[0]

	if !yes {
		confirmed, err := ui.Confirm(fmt.Sprintf("Are you sure you want to delete zone %s?", zoneIdentifier))
		if err != nil || !confirmed {
			fmt.Println(ui.Warning("Zone deletion cancelled."))
			return
		}
	}

	zoneID, err := cloudflare.ZoneIDByName(client, zoneIdentifier)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error finding zone", err))
		return
	}

	_, err = client.Zones.Delete(context.Background(), zones.ZoneDeleteParams{ZoneID: cf.F(zoneID)})
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error deleting zone", err))
		return
	}

	fmt.Println(ui.Success(fmt.Sprintf("Successfully deleted zone %s", zoneIdentifier)))
}
