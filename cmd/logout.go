package cmd

import (
	"errors"
	"os"

	"github.com/dajooo/cloudflare-cli/internal/config"
	"github.com/dajooo/cloudflare-cli/internal/ui"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout",
	Run:   executeLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func executeLogout(cmd *cobra.Command, args []string) {
	err := config.LoadConfig()
	if err != nil {
		if errors.Is(err, config.ErrNotLoggedIn) {
			println(ui.ErrorBox("You are not logged in."))
			return
		}
		println(ui.ErrorBox("Error loading config.", err))
		os.Exit(1)
	}
	config.Cfg.APIToken = ""
	config.Cfg.APIEmail = ""
	config.Cfg.APIKey = ""
	err = config.SaveConfig()
	if err != nil {
		println(ui.ErrorBox("Error saving config.", err))
		os.Exit(1)
	}
	println(ui.Success("You were successfully logged out."))
}
