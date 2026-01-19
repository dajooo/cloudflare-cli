package flags

import "github.com/spf13/cobra"

const AccountIDFlag = "account-id"

func RegisterAccountID(cmd *cobra.Command) {
	cmd.PersistentFlags().String(AccountIDFlag, "", "Account ID to use for this command")
}
