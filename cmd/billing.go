package cmd

import (
	"context"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/flags"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/billing"
	"github.com/spf13/cobra"
)

var billingCmd = &cobra.Command{
	Use:   "billing",
	Short: "Manage billing information",
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage billing profile",
}

var billingProfileGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get billing profile",
	Run: executor.NewBuilder[*cf.Client, *billing.ProfileGetResponse]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching billing profile", func(client *cf.Client, cmd *cobra.Command, args []string, progress chan<- string) (*billing.ProfileGetResponse, error) {
			accountID, err := cloudflare.GetAccountID(client, cmd, "")
			if err != nil {
				return nil, err
			}

			profile, err := client.Billing.Profiles.Get(context.Background(), billing.ProfileGetParams{
				AccountID: cf.F(accountID),
			})
			if err != nil {
				return nil, err
			}
			return profile, nil
		}).
		Display(func(profile *billing.ProfileGetResponse, fetchDuration time.Duration, err error) {
			rb := response.New().Title("Billing Profile")
			if err != nil {
				rb.Error("Error fetching billing profile", err).Display()
				return
			}

			icb := response.NewItemContent().
				Add("First Name:", ui.Text(profile.FirstName)).
				Add("Last Name:", ui.Text(profile.LastName)).
				Add("Company:", ui.Text(profile.Company)).
				Add("Address:", ui.Text(profile.Address)).
				Add("City:", ui.Text(profile.City)).
				Add("State:", ui.Text(profile.State)).
				Add("Zip:", ui.Text(profile.Zipcode)).
				Add("Country:", ui.Text(profile.Country))

			if profile.Telephone != "" {
				icb.Add("Phone:", ui.Text(profile.Telephone))
			}

			if profile.PaymentEmail != "" {
				icb.Add("Payment Email:", ui.Text(profile.PaymentEmail))
			}

			rb.AddItem("Profile Details", icb.String())
			rb.FooterSuccess("Fetched billing profile in %v", fetchDuration)
			rb.Display()
		}).
		Build().
		CobraRun(),
}

func init() {
	flags.RegisterAccountID(billingCmd)
	profileCmd.AddCommand(billingProfileGetCmd)
	billingCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(billingCmd)
}
