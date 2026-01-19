package zone

import (
	"dario.lol/cf/internal/flags"
	"github.com/spf13/cobra"
)

var ZoneCmd = &cobra.Command{
	Use:   "zone",
	Short: "Manage Cloudflare zones",
}

func init() {
	flags.RegisterAccountID(ZoneCmd)
}
