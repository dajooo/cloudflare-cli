package account

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/spf13/cobra"
)

var accountsKey = executor.NewKey[[]accounts.Account]("accounts")

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accessible accounts",
	Run: executor.New().
		WithClient().
		WithPagination().
		WithNoCache().
		Step(executor.NewStep(accountsKey, "Fetching accounts").
			Func(fetchAccounts).
			CacheKey("accounts:list")).
		Display(printAccountsList).
		Run(),
}

func init() {
	pagination.RegisterFlags(listCmd)
	listCmd.Flags().Bool("no-cache", false, "Don't use the cache when listing")
	AccountCmd.AddCommand(listCmd)
}

func fetchAccounts(ctx *executor.Context, _ chan<- string) ([]accounts.Account, error) {
	accountsList, err := ctx.Client.Accounts.List(context.Background(), accounts.AccountListParams{})
	if err != nil {
		return nil, err
	}
	return accountsList.Result, nil
}

func printAccountsList(ctx *executor.Context) {
	rb := response.New().
		Title("Accessible Accounts").
		NoItemsMessage("No accounts found")

	if ctx.Error != nil {
		rb.Error("Error fetching accounts", ctx.Error).Display()
		return
	}

	accountsList := executor.Get(ctx, accountsKey)
	paginated, info := pagination.Paginate(accountsList, ctx.Pagination)

	rb.Summary("Total:", info.Total)

	currentAccountID := config.Cfg.AccountID

	for i, account := range paginated {
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

	if len(paginated) > 0 {
		footer := info.FooterMessage("account(s)")
		footer += " " + ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))
		rb.FooterSuccess(footer)
	}

	rb.Display()
}
