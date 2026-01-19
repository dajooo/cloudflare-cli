package zone

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cobra"
)

var createdZoneKey = executor.NewKey[*zones.Zone]("createdZone")

var createCmd = &cobra.Command{
	Use:   "create <domain>",
	Short: "Adds a new domain to Cloudflare",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(createdZoneKey, "Creating zone").Func(createZone)).
		Display(printCreateZoneResult).
		Run(),
}

func init() {
	createCmd.Flags().Bool("jumpstart", false, "Automatically scan for common DNS records")
	ZoneCmd.AddCommand(createCmd)
}

func createZone(ctx *executor.Context, _ chan<- string) (*zones.Zone, error) {
	domain := ctx.Args[0]
	params := zones.ZoneNewParams{
		Name:    cf.F(domain),
		Account: cf.F(zones.ZoneNewParamsAccount{ID: cf.F(ctx.AccountID)}),
	}

	zone, err := ctx.Client.Zones.New(context.Background(), params)
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func printCreateZoneResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error creating zone", ctx.Error).Display()
		return
	}
	zone := executor.Get(ctx, createdZoneKey)
	rb.FooterSuccessf("Successfully created zone %s (%s) %s", zone.Name, zone.ID, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
