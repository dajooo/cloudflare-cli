package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/ui"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var migrateCmd = &cobra.Command{
	Use:    "migrate",
	Short:  "Migrates legacy config and cache files to the new database format",
	Hidden: true,
	Run:    executeMigrate,
}

type oldCacheEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}
type oldCacheData struct {
	Identifiers map[string]oldCacheEntry `json:"identifiers"`
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func executeMigrate(cmd *cobra.Command, args []string) {
	fmt.Println(ui.Title("Starting Migration"))
	fmt.Println(ui.Text("Migrating legacy config files (.yaml, .json) to the new database format..."))
	fmt.Println()

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(ui.ErrorBox("Could not determine home directory", err))
		os.Exit(1)
	}
	configDir := filepath.Join(home, ".cloudflare-cli")

	configPath := filepath.Join(configDir, "config.yaml")
	cachePath := filepath.Join(configDir, "cache.json")

	migratedSomething := false

	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		fmt.Println(ui.Info("Found legacy config.yaml, migrating configuration..."))
		if err := migrateConfig(configDir); err != nil {
			fmt.Println(ui.ErrorBox("Failed to migrate configuration", err))
		} else {
			fmt.Println(ui.Success("Configuration migrated successfully."))
			migratedSomething = true
		}
		fmt.Println()
	}

	if _, err := os.Stat(cachePath); !errors.Is(err, os.ErrNotExist) {
		fmt.Println(ui.Info("Found legacy cache.json, migrating cache..."))
		if err := migrateCache(cachePath); err != nil {
			fmt.Println(ui.ErrorBox("Failed to migrate cache", err))
		} else {
			fmt.Println(ui.Success("Cache migrated successfully."))
			migratedSomething = true
		}
		fmt.Println()
	}

	if !migratedSomething {
		fmt.Println(ui.Warning("No legacy configuration or cache files were found."))
		return
	}

	fmt.Println(ui.Title("Migration Complete!"))
	fmt.Println(ui.Text("You can now safely delete the following files:"))
	fmt.Println(ui.Muted(configPath))
	fmt.Println(ui.Muted(cachePath))
}

func migrateConfig(configPath string) error {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("could not read legacy config file: %w", err)
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return fmt.Errorf("could not unmarshal legacy config: %w", err)
	}

	config.Cfg = cfg
	return config.SaveConfig()
}

func migrateCache(cachePath string) error {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return fmt.Errorf("could not read legacy cache file: %w", err)
	}

	var cacheData oldCacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return fmt.Errorf("could not unmarshal legacy cache data: %w", err)
	}

	if len(cacheData.Identifiers) == 0 {
		return nil
	}

	count := 0
	for key, entry := range cacheData.Identifiers {
		cloudflare.SetID(key, entry.ID)
		count++
	}
	fmt.Printf("  Migrated %d cache entries.\n", count)
	return nil
}
