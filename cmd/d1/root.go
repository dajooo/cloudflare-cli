package d1

import (
	"dario.lol/cf/internal/flags"
	"github.com/spf13/cobra"
)

var D1Cmd = &cobra.Command{
	Use:   "d1",
	Short: "Manage D1 databases",
}

func init() {
	flags.RegisterAccountID(D1Cmd)
}
