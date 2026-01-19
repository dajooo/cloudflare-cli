package cmd

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/flags"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/billing"
	"github.com/spf13/cobra"
)

var billingProfileKey = executor.NewKey[*billing.ProfileGetResponse]("billingProfile")

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
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(billingProfileKey, "Fetching billing profile").Func(fetchBillingProfile)).
		Display(printBillingProfile).
		Run(),
}

func init() {
	flags.RegisterAccountID(billingCmd)
	profileCmd.AddCommand(billingProfileGetCmd)
	billingCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(billingCmd)
}

func fetchBillingProfile(ctx *executor.Context, _ chan<- string) (*billing.ProfileGetResponse, error) {
	return ctx.Client.Billing.Profiles.Get(context.Background(), billing.ProfileGetParams{
		AccountID: cf.F(ctx.AccountID),
	})
}

func printBillingProfile(ctx *executor.Context) {
	rb := response.New().Title("Billing Profile")

	if ctx.Error != nil {
		rb.Error("Error fetching billing profile", ctx.Error).Display()
		return
	}

	profile := executor.Get(ctx, billingProfileKey)

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
	rb.FooterSuccessf("Fetched billing profile %s", ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration)))
	rb.Display()
}
