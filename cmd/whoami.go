package cmd

import (
	"context"
	"fmt"
	"strings"

	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	"github.com/cloudflare/cloudflare-go/v6/user"
	"github.com/spf13/cobra"
)

var userKey = executor.NewKey[*user.UserGetResponse]("user")

var whoAmICmd = &cobra.Command{
	Use:   "whoami",
	Short: "Get current user",
	Run: executor.New().
		WithClient().
		WithNoCache().
		Step(executor.NewStep(userKey, "Fetching user information").
			Func(fetchUser).
			CacheKey("user:whoami")).
		Display(printUserInfo).
		Run(),
}

func init() {
	whoAmICmd.Flags().Bool("no-cache", false, "Bypass the cache and fetch directly from the API")
	rootCmd.AddCommand(whoAmICmd)
}

func fetchUser(ctx *executor.Context, _ chan<- string) (*user.UserGetResponse, error) {
	return ctx.Client.User.Get(context.Background())
}

func printUserInfo(ctx *executor.Context) {
	rb := response.New().Title("Account Information")

	if ctx.Error != nil {
		rb.Error("Error getting user information", ctx.Error).Display()
		return
	}

	user := executor.Get(ctx, userKey)

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

	contextContent := response.NewItemContent()
	if config.Cfg.AccountID != "" {
		contextContent.Add("Account ID:", ui.Text(config.Cfg.AccountID))
	} else {
		contextContent.Add("Account ID:", ui.Muted("Not selected"))
	}
	if config.Cfg.KVNamespaceID != "" {
		contextContent.Add("KV Namespace:", ui.Text(config.Cfg.KVNamespaceID))
	} else {
		contextContent.Add("KV Namespace:", ui.Muted("Not selected"))
	}
	rb.AddItem("Current Context", contextContent.String())

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
		copy(betaItems, user.Betas)
		rb.AddItem("Beta Programs", ui.BulletList(betaItems))
	}

	rb.FooterSuccessf("Authentication successful %s", ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).
		Display()
}
