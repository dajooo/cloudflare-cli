package cmd

import (
	"context"
	"os"

	"dario.lol/cf/internal/ui"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cf",
	Short: "CLI to control Cloudflare",
	Long:  ``,
}

func configureColorScheme(_ lipgloss.LightDarkFunc) fang.ColorScheme {
	return ui.FangTheme()
}

func Execute() {
	if err := fang.Execute(context.Background(), rootCmd, fang.WithColorSchemeFunc(configureColorScheme)); err != nil {
		println(ui.ErrorBox("Error executing command", err))
		os.Exit(1)
	}
}
