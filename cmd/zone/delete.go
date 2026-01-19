package zone

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
)

type DeletedZoneInfo struct {
	ID   string
	Name string
}

var deletedZoneKey = executor.NewKey[*DeletedZoneInfo]("deletedZone")

var deleteCmd = &cobra.Command{
	Use:   "delete <zone_name|zone_id>",
	Short: "Deletes a zone",
	Args:  cobra.ExactArgs(1),
	Run:   executeZoneDelete,
}

func init() {
	deleteCmd.Flags().Bool("yes", false, "Bypass the confirmation prompt")
	ZoneCmd.AddCommand(deleteCmd)
}

func executeZoneDelete(cmd *cobra.Command, args []string) {
	zoneIdentifier := args[0]
	yes, _ := cmd.Flags().GetBool("yes")

	if !yes {
		confirmed, err := ui.Confirm(fmt.Sprintf("Are you sure you want to delete zone %s?", zoneIdentifier))
		if err != nil || !confirmed {
			fmt.Println(ui.Warning("Zone deletion cancelled."))
			return
		}
	}

	executor.New().
		WithClient().
		Step(executor.NewStep(deletedZoneKey, "Deleting zone").Func(deleteZone)).
		Display(printDeleteZoneResult).
		Run()(cmd, args)
}

func deleteZone(ctx *executor.Context, _ chan<- string) (*DeletedZoneInfo, error) {
	zoneIdentifier := ctx.Args[0]

	zoneID, zoneName, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %w", err)
	}

	_, err = ctx.Client.Zones.Delete(context.Background(), zones.ZoneDeleteParams{ZoneID: cf.F(zoneID)})
	if err != nil {
		return nil, err
	}

	return &DeletedZoneInfo{
		ID:   zoneID,
		Name: zoneName,
	}, nil
}

func printDeleteZoneResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error deleting zone", ctx.Error).Display()
		return
	}
	zone := executor.Get(ctx, deletedZoneKey)
	rb.FooterSuccessf("Successfully deleted zone %s (%s) %s", zone.Name, zone.ID, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
