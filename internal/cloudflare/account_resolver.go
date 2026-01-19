package cloudflare

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/flags"
	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/spf13/cobra"
)

func GetAccountID(client *cloudflare.Client, cmd *cobra.Command, accountID string) (string, error) {
	if accountID != "" {
		return accountID, nil
	}

	if cmd != nil {
		if flagID, _ := cmd.Flags().GetString(flags.AccountIDFlag); flagID != "" {
			return flagID, nil
		}
	}

	if config.Cfg.AccountID != "" {
		return config.Cfg.AccountID, nil
	}

	list, err := client.Accounts.List(context.Background(), accounts.AccountListParams{})
	if err != nil {
		return "", fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(list.Result) == 0 {
		return "", fmt.Errorf("no accounts found")
	}

	if len(list.Result) == 1 {
		return list.Result[0].ID, nil
	}

	return "", fmt.Errorf("multiple accounts found, please specify --account-id or use `cf account switch`")
}
