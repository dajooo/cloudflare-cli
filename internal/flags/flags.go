package flags

import "github.com/spf13/cobra"

const (
	AccountIDFlag = "account-id"
	YesFlag       = "yes"
)

func RegisterAccountID(cmd *cobra.Command) {
	cmd.PersistentFlags().String(AccountIDFlag, "", "Account ID to use for this command")
}

func RegisterConfirmation(cmd *cobra.Command) {
	cmd.Flags().BoolP(YesFlag, "y", false, "Skip confirmation prompt")
}
