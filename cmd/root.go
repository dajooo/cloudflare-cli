package cmd

import (
	"context"
	"fmt"
	"os"

	"dario.lol/cf/internal/constants"
	"dario.lol/cf/internal/ui"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "cf",
	Short:   fmt.Sprintf("CLI to control Cloudflare version %s", constants.Version),
	Long:    ``,
	Version: constants.Version,
}

func configureColorScheme(_ lipgloss.LightDarkFunc) fang.ColorScheme {
	return ui.FangTheme()
}

func Execute() {
	if err := fang.Execute(context.Background(), rootCmd, fang.WithColorSchemeFunc(configureColorScheme), fang.WithVersion(constants.Version)); err != nil {
		println(ui.ErrorBox("Error executing command", err))
		os.Exit(1)
	}
}
