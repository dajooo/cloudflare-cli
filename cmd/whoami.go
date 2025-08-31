package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/user"
	"github.com/spf13/cobra"
)

var whoAmICmd = &cobra.Command{
	Use:   "whoami",
	Short: "Get current user",
	Run: executor.NewBuilder[*cf.Client, *user.UserGetResponse]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching user information", fetchUser).
		Display(printUserInfo).
		Build().
		CobraRun(),
}

func init() {
	rootCmd.AddCommand(whoAmICmd)
}

func fetchUser(client *cf.Client, _ *cobra.Command, _ []string, _ chan<- string) (*user.UserGetResponse, error) {
	return client.User.Get(context.Background())
}

func printUserInfo(user *user.UserGetResponse, fetchDuration time.Duration, err error) {
	rb := response.New().Title("Account Information")

	if err != nil {
		rb.Error("Error getting user information", err).Display()
		return
	}

	identityContent := response.NewItemContent()
	if user.FirstName != "" || user.LastName != "" {
		if fullName := strings.TrimSpace(user.FirstName + " " + user.LastName); fullName != "" {
			identityContent.Add("Name:", ui.Text(fullName))
		}
	}
	identityContent.Add("User ID:", ui.Muted(user.ID))
	if user.Country != "" {
		identityContent.Add("Country:", ui.Text(user.Country))
	}
	rb.AddItem("User Identity", identityContent.String())

	statusContent := response.NewItemContent()
	if user.Suspended {
		statusContent.Add("Status:", ui.Error("Suspended"))
	} else {
		statusContent.Add("Status:", ui.Success("Active"))
	}
	if user.TwoFactorAuthenticationEnabled {
		statusContent.Add("2FA:", ui.Success("Enabled"))
	} else {
		statusContent.Add("2FA:", ui.Warning("Disabled"))
	}
	rb.AddItem("Account Status", statusContent.String())

	var zones []string
	if user.HasEnterpriseZones {
		zones = append(zones, ui.BadgeSuccess.Render("Enterprise"))
	}
	if user.HasBusinessZones {
		zones = append(zones, ui.BadgePrimary.Render("Business"))
	}
	if user.HasProZones {
		zones = append(zones, ui.Badge.Render("Pro"))
	}
	if len(zones) == 0 {
		zones = append(zones, ui.Muted("Free"))
	}
	zoneContent := response.NewItemContent().Add("Types:", strings.Join(zones, " "))
	rb.AddItem("Zone Information", zoneContent.String())

	if len(user.Organizations) > 0 {
		var orgContent strings.Builder
		for i, org := range user.Organizations {
			orgContent.WriteString(fmt.Sprintf("%s %s",
				ui.Muted(fmt.Sprintf("%d.", i+1)),
				ui.Text(org.Name)))
			if org.ID != "" {
				orgContent.WriteString(fmt.Sprintf("\n   %s", ui.Small("ID: "+org.ID)))
			}
			if i < len(user.Organizations)-1 {
				orgContent.WriteString("\n\n")
			}
		}
		rb.AddItem("Organizations", orgContent.String())
	}

	if len(user.Betas) > 0 {
		betaItems := make([]string, len(user.Betas))
		for i, beta := range user.Betas {
			betaItems[i] = beta
		}
		rb.AddItem("Beta Programs", ui.BulletList(betaItems))
	}

	rb.FooterSuccess("Authentication successful (took %v)", fetchDuration).
		Display()
}
