package zone

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
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
		Display(func(zone *zones.Zone, duration time.Duration, err error) {
			if err != nil {
				fmt.Println(ui.ErrorMessage("Error creating zone", err))
				return
			}
			fmt.Println(ui.Success(fmt.Sprintf("Successfully created zone %s (%s) in %v", zone.Name, zone.ID, duration)))
		}).
		Build().
		CobraRun(),
}

func init() {
	createCmd.Flags().StringVar(&createAccountID, "account-id", "", "The account ID to create the zone in")
	createCmd.Flags().BoolVar(&jumpstart, "jumpstart", false, "Automatically scan for common DNS records")
	ZoneCmd.AddCommand(createCmd)
}

func createZone(client *cf.Client, _ *cobra.Command, args []string) (*zones.Zone, error) {
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
