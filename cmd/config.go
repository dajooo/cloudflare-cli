package cmd

import (
	"fmt"
	"strings"

	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value (e.g., caching true|false)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])
		value := strings.ToLower(args[1])
		rb := response.New()
		err := config.LoadConfig()
		if err != nil {
			rb.Error("Error loading config", err)
			return
		}

		switch key {
		case "caching":
			if value == "true" || value == "1" || value == "on" {
				config.Cfg.Caching = true
			} else if value == "false" || value == "0" || value == "off" {
				config.Cfg.Caching = false
			} else {
				rb.Error("Invalid value", fmt.Errorf("caching must be true or false")).Display()
				return
			}
			if err := config.SaveConfig(); err != nil {
				rb.Error("Failed to save config", err).Display()
				return
			}
			rb.FooterSuccessf("Configuration updated: %s set to %v", ui.Code.Render(key), config.Cfg.Caching).Display()
		default:
			rb.Error("Unknown configuration key", fmt.Errorf("key %q is not supported", key)).Display()
		}
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])
		rb := response.New()

		switch key {
		case "caching":
			rb.FooterSuccessf("Current %s setting: %v", ui.Code.Render(key), config.Cfg.Caching).Display()
		default:
			rb.Error("Unknown configuration key", fmt.Errorf("key %q is not supported", key)).Display()
		}
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}
