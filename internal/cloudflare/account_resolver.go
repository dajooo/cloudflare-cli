package cloudflare

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
)

func GetAccountID(client *cloudflare.Client, accountID string) (string, error) {
	if accountID != "" {
		return accountID, nil
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

	return "", fmt.Errorf("multiple accounts found, please specify --account-id")
}
