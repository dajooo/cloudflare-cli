package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/dajooo/cloudflare-cli/internal/config"
	"github.com/dajooo/cloudflare-cli/internal/prompt"
	"github.com/dajooo/cloudflare-cli/internal/ui"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to cloudflare-cli",
	Run:   executeLogin,
}

var (
	loginEmail  string
	loginToken  string
	loginApiKey string
)

func init() {
	loginCmd.Flags().StringVarP(&loginEmail, "email", "e", "", "Cloudflare email")
	loginCmd.Flags().StringVarP(&loginToken, "token", "t", "", "Cloudflare token")
	loginCmd.Flags().StringVarP(&loginApiKey, "apikey", "k", "", "Cloudflare api key")

	rootCmd.AddCommand(loginCmd)
}

func executeLogin(cmd *cobra.Command, args []string) {
	if handleArgs() {
		println(ui.Success("You were successfully logged in."))
		return
	}
	handleInput()
}

func checkCredentials(token, email, apiKey string) error {
	var client *cloudflare.Client
	if token != "" {
		client = cloudflare.NewClient(option.WithAPIToken(token))
	}
	if email != "" && apiKey != "" {
		client = cloudflare.NewClient(option.WithAPIEmail(email), option.WithAPIKey(apiKey))
	}
	if client == nil {
		return errors.New("error creating cloudflare client")
	}
	me, err := client.User.Get(context.Background())
	if err != nil {
		return err
	}
	if me == nil {
		return errors.New("error getting user")
	}
	return nil
}

func handleArgs() bool {
	if loginToken != "" {
		err := checkCredentials(loginToken, "", "")
		config.Cfg.APIToken = config.EncryptedString(loginToken)
		err = config.SaveConfig()
		if err != nil {
			println(ui.ErrorBox("Error saving config.", err))
			os.Exit(1)
		}
		return true
	}
	if loginApiKey != "" || loginEmail != "" {
		if loginApiKey == "" {
			println(ui.ErrorBox("Cloudflare api key is required"))
			os.Exit(1)
		}
		if loginEmail == "" {
			println(ui.ErrorBox("Cloudflare email is required"))
			os.Exit(1)
		}
		err := checkCredentials("", loginEmail, loginApiKey)
		if err == nil {
			config.Cfg.APIEmail = loginEmail
			config.Cfg.APIKey = config.EncryptedString(loginApiKey)
			err := config.SaveConfig()
			if err != nil {
				println(ui.ErrorBox("Error loading config.", err))
				os.Exit(1)
			}
			return true
		}
	}
	return false
}

func handleInput() {
	credentials, err := prompt.RunLoginPrompt()
	if err != nil {
		if errors.Is(err, prompt.ErrUserCancelled) {
			return
		}
		println(ui.ErrorBox("Error reading login credentials.", err))
		os.Exit(1)
	}
	err = checkCredentials(credentials.Token, credentials.Email, credentials.APIKey)
	if err != nil {
		println(ui.ErrorBox("Invalid credentials, could not log in.", err))
		os.Exit(1)
	}
	if credentials.AuthMethod == prompt.TokenAuthMethod {
		config.Cfg.APIToken = config.EncryptedString(credentials.Token)
		err := config.SaveConfig()
		if err != nil {
			println(ui.ErrorBox("Error saving config.", err))
			os.Exit(1)
		}
		println(ui.Success("You were successfully logged in."))
	}
	if credentials.AuthMethod == prompt.LegacyAuthMethod {
		config.Cfg.APIEmail = credentials.Email
		config.Cfg.APIKey = config.EncryptedString(credentials.APIKey)
		err := config.SaveConfig()
		if err != nil {
			println(ui.ErrorBox("Error saving config.", err))
			os.Exit(1)
		}
		println(ui.Success("You were successfully logged in."))
	}
}
