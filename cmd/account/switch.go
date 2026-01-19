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

var switchCmd = &cobra.Command{
	Use:   "switch <name-or-id>",
	Short: "Switch the active account context",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, accounts.Account]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Verifying account", runAccountSwitch).
		Display(printAccountSwitch).
		Build().
		CobraRun(),
}

func runAccountSwitch(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (accounts.Account, error) {
	input := args[0]
	var selectedAccount accounts.Account

	list, err := client.Accounts.List(context.Background(), accounts.AccountListParams{
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
		account, err := client.Accounts.Get(context.Background(), accounts.AccountGetParams{
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

func printAccountSwitch(account accounts.Account, _ time.Duration, err error) {
	if err != nil {
		response.New().Error("Failed to switch account", err).Display()
		return
	}
	response.New().
		Title("Account Switched").
		FooterSuccess("Switched context to account %s (%s)", ui.Text(account.Name), ui.Muted(account.ID)).
		Display()
}

func init() {
	AccountCmd.AddCommand(switchCmd)
}
