package zone

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
)

var zonesKey = executor.NewKey[[]zones.Zone]("zones")

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones in an account",
	Run: executor.New().
		WithClient().
		WithAccountID().
		WithPagination().
		WithNoCache().
		Step(executor.NewStep(zonesKey, "Fetching zones").Func(fetchZones)).
		Caches(func(ctx *executor.Context) ([]string, error) {
			return []string{"zones:list"}, nil
		}).
		Display(printZonesList).
		Run(),
}

func init() {
	pagination.RegisterFlags(listCmd)
	listCmd.Flags().String("status", "", "The status of the zones to list (active, pending)")
	listCmd.Flags().Bool("no-cache", false, "Don't use the cache when listing records")
	ZoneCmd.AddCommand(listCmd)
}

func fetchZones(ctx *executor.Context, _ chan<- string) ([]zones.Zone, error) {
	params := zones.ZoneListParams{}
	params.Account.Value.ID.Value = ctx.AccountID

	if status, _ := ctx.Cmd.Flags().GetString("status"); status != "" {
		params.Status.Value = zones.ZoneListParamsStatus(status)
	}

	zonesList, err := ctx.Client.Zones.List(context.Background(), params)
	if err != nil {
		return nil, err
	}

	for _, zone := range zonesList.Result {
		cloudflare.SetID(cloudflare.ZoneCacheKey(zone.Name), zone.ID)
		cloudflare.SetID(cloudflare.ZoneCacheKey(zone.ID), zone.Name)
	}

	return zonesList.Result, nil
}

func printZonesList(ctx *executor.Context) {
	rb := response.New().
		Title("Accessible Zones").
		NoItemsMessage("No zones found")

	if ctx.Error != nil {
		rb.Error("Error fetching zones", ctx.Error).Display()
		return
	}

	zonesList := executor.Get(ctx, zonesKey)
	paginated, info := pagination.Paginate(zonesList, ctx.Pagination)

	rb.Summary("Total:", info.Total)

	for i, zone := range paginated {
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

	if len(paginated) > 0 {
		footer := fmt.Sprintf("Showing %d of %d zone(s)", info.Showing, info.Total)
		if info.HasMore {
			footer += fmt.Sprintf(" (page %d)", info.Page)
		}
		footer += " " + ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))
		rb.FooterSuccess(footer)
	}

	rb.Display()
}
