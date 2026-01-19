package account

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/spf13/cobra"
)

var switchedAccountKey = executor.NewKey[accounts.Account]("switchedAccount")

var switchCmd = &cobra.Command{
	Use:   "switch <name-or-id>",
	Short: "Switch the active account context",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		Step(executor.NewStep(switchedAccountKey, "Verifying account").Func(runAccountSwitch)).
		Display(printAccountSwitch).
		Run(),
}

func init() {
	AccountCmd.AddCommand(switchCmd)
}

func runAccountSwitch(ctx *executor.Context, _ chan<- string) (accounts.Account, error) {
	input := ctx.Args[0]
	var selectedAccount accounts.Account

	list, err := ctx.Client.Accounts.List(context.Background(), accounts.AccountListParams{
		Name: cf.F(input),
	})
	if err != nil {
		return accounts.Account{}, fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(list.Result) == 1 {
		selectedAccount = list.Result[0]
	} else if len(list.Result) > 1 {
		msg := fmt.Sprintf("Multiple accounts found with name '%s':", input)
		for _, acc := range list.Result {
			msg += fmt.Sprintf("\n - %s (%s)", acc.Name, acc.ID)
		}
		return accounts.Account{}, fmt.Errorf("%s\nPlease specify the Account ID.", msg)
	} else {
		account, err := ctx.Client.Accounts.Get(context.Background(), accounts.AccountGetParams{
			AccountID: cf.F(input),
		})
		if err != nil {
			return accounts.Account{}, fmt.Errorf("account not found with name or ID '%s'", input)
		}
		selectedAccount = *account
	}

	config.Cfg.AccountID = selectedAccount.ID
	if err := config.SaveConfig(); err != nil {
		return accounts.Account{}, fmt.Errorf("failed to save config: %w", err)
	}

	return selectedAccount, nil
}

func printAccountSwitch(ctx *executor.Context) {
	if ctx.Error != nil {
		response.New().Error("Failed to switch account", ctx.Error).Display()
		return
	}
	account := executor.Get(ctx, switchedAccountKey)
	response.New().
		Title("Account Switched").
		FooterSuccessf("Switched to account %s (%s) %s", account.Name, account.ID, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).
		Display()
}
