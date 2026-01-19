package zone

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/flags"
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
	flags.RegisterConfirmation(deleteCmd)
	ZoneCmd.AddCommand(deleteCmd)
}

func executeZoneDelete(cmd *cobra.Command, args []string) {
	executor.New().
		WithClient().
		WithZone().
		WithConfirmationFunc(func(ctx *executor.Context) string {
			return fmt.Sprintf("Are you sure you want to delete zone %s (%s)?", executor.Get(ctx, executor.ZoneNameKey), executor.Get(ctx, executor.ZoneIDKey))
		}).
		Step(executor.NewStep(deletedZoneKey, "Deleting zone").Func(deleteZone)).
		Invalidates(func(ctx *executor.Context) []string {
			return []string{"zones:list", "zone:" + executor.Get(ctx, executor.ZoneIDKey) + ":"}
		}).
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
