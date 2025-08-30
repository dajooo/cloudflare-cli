package zone

import (
	"github.com/spf13/cobra"
)

var ZoneCmd = &cobra.Command{
	Use:   "zone",
	Short: "Manage Cloudflare zones",
}
