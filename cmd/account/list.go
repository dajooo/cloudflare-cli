package account

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/spf13/cobra"
)

var noCache bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accessible accounts",
	Run: executor.NewBuilder[*cf.Client, []accounts.Account]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching accounts", fetchAccounts).
		SkipCache(noCache).
		Caches(func(cmd *cobra.Command, args []string) ([]string, error) {
			return []string{"accounts:list"}, nil
		}).
		Display(printAccountsList).
		Build().
		CobraRun(),
}

func init() {
	listCmd.Flags().BoolVar(&noCache, "no-cache", false, "Don't use the cache when listing records")
	AccountCmd.AddCommand(listCmd)
}

func fetchAccounts(client *cf.Client, _ *cobra.Command, _ []string, _ chan<- string) ([]accounts.Account, error) {
	accountsList, err := client.Accounts.List(context.Background(), accounts.AccountListParams{})
	if err != nil {
		return nil, err
	}
	return accountsList.Result, nil
}

func printAccountsList(accountsList []accounts.Account, fetchDuration time.Duration, err error) {
	rb := response.New().
		Title("Accessible Accounts").
		NoItemsMessage("No accounts found")

	if err != nil {
		rb.Error("Error fetching accounts", err).Display()
		return
	}

	rb.Summary("Total:", len(accountsList))

	currentAccountID := config.Cfg.AccountID

	for i, account := range accountsList {
		icb := response.NewItemContent().
			Add("Name:", ui.Text(account.Name))

		if account.ID == currentAccountID {
			icb.Add("ID:", ui.Success(account.ID+" (Active)"))
		} else {
			icb.Add("ID:", ui.Muted(account.ID))
		}

		if !account.CreatedOn.IsZero() {
			icb.Add("Created:", ui.Small(account.CreatedOn.Format("2006-01-02 15:04:05")))
		}

		icb.AddRaw("\n" + ui.Subtitle("Settings:"))
		if account.Settings.AbuseContactEmail == "" {
			icb.Add("Abuse Email:", ui.Error("Not set"))
		} else {
			icb.Add("Abuse Email:", ui.Small(account.Settings.AbuseContactEmail))
		}

		if account.Settings.EnforceTwofactor {
			icb.Add("2FA Mode:", ui.Success("Enforced"))
		} else {
			icb.Add("2FA Mode:", ui.Error("Not enforced"))
		}

		cardTitle := fmt.Sprintf("Account %d: %s", i+1, account.Name)
		rb.AddItem(cardTitle, icb.String())
	}

	if len(accountsList) > 0 {
		rb.FooterSuccess("Found %d accessible account(s) %s", len(accountsList), ui.Muted(fmt.Sprintf("(took %v)", fetchDuration)))
	}

	rb.Display()
}
