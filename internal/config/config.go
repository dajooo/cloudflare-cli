package config

import (
	"errors"
	"os"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	APIToken EncryptedString `mapstructure:"api_token"`
	APIKey   EncryptedString `mapstructure:"api_key"`
	APIEmail string          `mapstructure:"api_email"`
}

var ErrNotLoggedIn = errors.New("either api_token or api_email and api_key must be set")

var Cfg Config

func initViper() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	viper.SetConfigName(".cloudflare-cli")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(home)

	viper.SetEnvPrefix("CF")
	viper.AutomaticEnv()
	_ = viper.BindEnv("api_token")
	_ = viper.BindEnv("api_key")
	_ = viper.BindEnv("api_email")

	return nil
}

func LoadConfig() error {
	if err := initViper(); err != nil {
		return err
	}

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return err
		}
	}

	if err := viper.Unmarshal(&Cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return err
	}

	if Cfg.APIToken == "" && Cfg.APIEmail == "" && Cfg.APIKey == "" {
		return ErrNotLoggedIn
	}
	return nil
}

func SaveConfig() error {
	if err := initViper(); err != nil {
		return err
	}

	viper.Set("api_token", Cfg.APIToken)
	viper.Set("api_key", Cfg.APIKey)
	viper.Set("api_email", Cfg.APIEmail)

	if err := viper.WriteConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return viper.SafeWriteConfig()
		}
		return err
	}
	return nil
}
