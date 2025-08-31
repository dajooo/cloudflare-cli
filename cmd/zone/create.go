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

var createAccountID string
var jumpstart bool

var createCmd = &cobra.Command{
	Use:   "create <domain>",
	Short: "Adds a new domain to Cloudflare",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, *zones.Zone]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Creating zone", createZone).
		Display(printCreateZoneResult).
		Build().
		CobraRun(),
}

func init() {
	createCmd.Flags().StringVar(&createAccountID, "account-id", "", "The account ID to create the zone in")
	createCmd.Flags().BoolVar(&jumpstart, "jumpstart", false, "Automatically scan for common DNS records")
	ZoneCmd.AddCommand(createCmd)
}

func createZone(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*zones.Zone, error) {
	domain := args[0]
	params := zones.ZoneNewParams{
		Name: cf.F(domain),
	}
	if createAccountID != "" {
		params.Account = cf.F(zones.ZoneNewParamsAccount{ID: cf.F(createAccountID)})
	}

	zone, err := client.Zones.New(context.Background(), params)
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func printCreateZoneResult(zone *zones.Zone, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error creating zone", err).Display()
		return
	}
	rb.FooterSuccess("Successfully created zone %s (%s) %s", zone.Name, zone.ID, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
