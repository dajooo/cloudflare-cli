package ssl

import (
	"github.com/spf13/cobra"
)

var SslCmd = &cobra.Command{
	Use:   "ssl",
	Short: "Manage SSL/TLS encryption modes",
}
