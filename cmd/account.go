package cmd

import (
	"context"
	"fmt"
	"strings"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/ui"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage Cloudflare accounts",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accessible accounts",
	Run:   executeAccountList,
}

func init() {
	accountCmd.AddCommand(listCmd)
	rootCmd.AddCommand(accountCmd)
}

func executeAccountList(cmd *cobra.Command, args []string) {
	client, err := cloudflare.NewClient()
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error loading config", err))
		return
	}

	accountsList, err := client.Accounts.List(context.Background(), accounts.AccountListParams{})
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error fetching accounts", err))
		return
	}

	printAccountsList(accountsList.Result)
}

func printAccountsList(accountsList []accounts.Account) {
	fmt.Println(ui.Title("Accessible Accounts"))
	fmt.Println()

	if len(accountsList) == 0 {
		fmt.Println(ui.Warning("No accounts found"))
		return
	}

	summaryContent := fmt.Sprintf("%-12s %d", "Total:", len(accountsList))
	fmt.Println(ui.Box(summaryContent, "Summary"))
	fmt.Println()

	for i, account := range accountsList {
		var accountContent strings.Builder

		accountContent.WriteString(fmt.Sprintf("%-12s %s\n", "Name:", ui.Text(account.Name)))
		accountContent.WriteString(fmt.Sprintf("%-12s %s", "ID:", ui.Muted(account.ID)))

		if !account.CreatedOn.IsZero() {
			accountContent.WriteString(fmt.Sprintf("\n%-12s %s", "Created:", ui.Small(account.CreatedOn.Format("2006-01-02 15:04:05"))))
		}

		accountContent.WriteString("\n\n")
		accountContent.WriteString(ui.Subtitle("Settings:"))

		if account.Settings.AbuseContactEmail == "" {
			accountContent.WriteString(fmt.Sprintf("\n%-12s %s", "Abuse Email:", ui.Error("Not set")))
		} else {
			accountContent.WriteString(fmt.Sprintf("\n%-12s %s", "Abuse Email:", ui.Small(account.Settings.AbuseContactEmail)))
		}

		if account.Settings.EnforceTwofactor {
			accountContent.WriteString(fmt.Sprintf("\n%-12s %s", "2FA Mode:", ui.Success("Enforced")))
		} else {
			accountContent.WriteString(fmt.Sprintf("\n%-12s %s", "2FA Mode:", ui.Error("Not enforced")))
		}

		cardTitle := fmt.Sprintf("Account %d", i+1)
		if account.Name != "" {
			cardTitle = fmt.Sprintf("Account %d: %s", i+1, account.Name)
			if len(cardTitle) > 40 {
				cardTitle = fmt.Sprintf("Account %d: %s...", i+1, account.Name[:25])
			}
		}

		fmt.Println(ui.Box(accountContent.String(), cardTitle))
		fmt.Println()
	}

	fmt.Println(ui.Success(fmt.Sprintf("Found %d accessible account(s)", len(accountsList))))
}
