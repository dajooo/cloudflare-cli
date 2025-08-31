package config

import (
	"errors"

	"dario.lol/cf/internal/db"
)

type Config struct {
	APIToken EncryptedString `mapstructure:"api_token"`
	APIKey   EncryptedString `mapstructure:"api_key"`
	APIEmail string          `mapstructure:"api_email"`
}

var ErrNotLoggedIn = errors.New("you are not logged in. Please use 'cf login'")

var Cfg Config

func LoadConfig() error {
	var newCfg Config

	token, err := db.Get(db.ConfigBucket, []byte("api_token"))
	if err != nil {
		return err
	}
	if err := newCfg.APIToken.UnmarshalText(token); err != nil {
		return err
	}

	key, err := db.Get(db.ConfigBucket, []byte("api_key"))
	if err != nil {
		return err
	}
	if err := newCfg.APIKey.UnmarshalText(key); err != nil {
		return err
	}

	email, err := db.Get(db.ConfigBucket, []byte("api_email"))
	if err != nil {
		return err
	}
	newCfg.APIEmail = string(email)

	Cfg = newCfg

	if Cfg.APIToken == "" && Cfg.APIEmail == "" && Cfg.APIKey == "" {
		return ErrNotLoggedIn
	}
	return nil
}

func SaveConfig() error {
	tokenBytes, err := Cfg.APIToken.MarshalText()
	if err != nil {
		return err
	}
	if err := db.Set(db.ConfigBucket, []byte("api_token"), tokenBytes); err != nil {
		return err
	}

	keyBytes, err := Cfg.APIKey.MarshalText()
	if err != nil {
		return err
	}
	if err := db.Set(db.ConfigBucket, []byte("api_key"), keyBytes); err != nil {
		return err
	}

	if err := db.Set(db.ConfigBucket, []byte("api_email"), []byte(Cfg.APIEmail)); err != nil {
		return err
	}

	return nil
}
