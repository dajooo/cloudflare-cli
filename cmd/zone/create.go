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

var createAccountID string
var jumpstart bool

var createCmd = &cobra.Command{
	Use:   "create <domain>",
	Short: "Adds a new domain to Cloudflare",
	Args:  cobra.ExactArgs(1),
	Run:   executeZoneCreate,
}

func init() {
	createCmd.Flags().StringVar(&createAccountID, "account-id", "", "The account ID to create the zone in")
	createCmd.Flags().BoolVar(&jumpstart, "jumpstart", false, "Automatically scan for common DNS records")
	ZoneCmd.AddCommand(createCmd)
}

func executeZoneCreate(cmd *cobra.Command, args []string) {
	client, err := cloudflare.NewClient()
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error loading config", err))
		return
	}

	domain := args[0]
	params := zones.ZoneNewParams{
		Name: cf.F(domain),
	}
	if createAccountID != "" {
		params.Account = cf.F(zones.ZoneNewParamsAccount{ID: cf.F(createAccountID)})
	}

	zone, err := client.Zones.New(context.Background(), params)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error creating zone", err))
		return
	}

	fmt.Println(ui.Success(fmt.Sprintf("Successfully created zone %s", zone.Name)))
}
