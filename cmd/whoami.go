package cmd

import (
	"context"
	"fmt"
	"strings"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/ui"
	"github.com/cloudflare/cloudflare-go/v6/user"
	"github.com/spf13/cobra"
)

var whoAmICmd = &cobra.Command{
	Use:   "whoami",
	Short: "Get current user",
	Run:   executeWhoAmI,
}

func init() {
	rootCmd.AddCommand(whoAmICmd)
}

func executeWhoAmI(cmd *cobra.Command, args []string) {
	client, err := cloudflare.NewClient()
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error loading config", err))
		return
	}

	me, err := client.User.Get(context.Background())
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error getting user information", err))
		return
	}

	printUserInfo(me)
}

func printUserInfo(user *user.UserGetResponse) {
	fmt.Println(ui.Title("Account Information"))
	fmt.Println()

	var identityContent strings.Builder

	if user.FirstName != "" || user.LastName != "" {
		fullName := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if fullName != "" {
			identityContent.WriteString(fmt.Sprintf("%-10s %s\n", "Name:", ui.Text(fullName)))
		}
	}

	identityContent.WriteString(fmt.Sprintf("%-10s %s", "User ID:", ui.Muted(user.ID)))

	if user.Country != "" {
		identityContent.WriteString(fmt.Sprintf("\n%-10s %s", "Country:", ui.Text(user.Country)))
	}

	fmt.Println(ui.Box(identityContent.String(), "User Identity"))
	fmt.Println()

	var statusContent strings.Builder

	if user.Suspended {
		statusContent.WriteString(fmt.Sprintf("%-10s %s\n", "Status:", ui.Error("Suspended")))
	} else {
		statusContent.WriteString(fmt.Sprintf("%-10s %s\n", "Status:", ui.Success("Active")))
	}

	if user.TwoFactorAuthenticationEnabled {
		statusContent.WriteString(fmt.Sprintf("%-10s %s", "2FA:", ui.Success("Enabled")))
	} else {
		statusContent.WriteString(fmt.Sprintf("%-10s %s", "2FA:", ui.Warning("Disabled")))
	}

	fmt.Println(ui.Box(statusContent.String(), "Account Status"))
	fmt.Println()

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

	zoneContent := fmt.Sprintf("%-10s %s", "Types:", strings.Join(zones, " "))
	fmt.Println(ui.Box(zoneContent, "Zone Information"))
	fmt.Println()

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

		fmt.Println(ui.Box(orgContent.String(), "Organizations"))
		fmt.Println()
	}

	if len(user.Betas) > 0 {
		betaItems := make([]string, len(user.Betas))
		for i, beta := range user.Betas {
			betaItems[i] = beta
		}

		betaContent := ui.BulletList(betaItems)
		fmt.Println(ui.Box(betaContent, "Beta Programs"))
		fmt.Println()
	}

	fmt.Println(ui.Success("Authentication successful"))
}
