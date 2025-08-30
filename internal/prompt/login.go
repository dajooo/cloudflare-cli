package prompt

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/dajooo/cloudflare-cli/internal/ui"
)

type AuthMethod string

var (
	TokenAuthMethod  AuthMethod = "token"
	LegacyAuthMethod AuthMethod = "legacy"
)

type Credentials struct {
	AuthMethod AuthMethod
	Token      string
	Email      string
	APIKey     string
}

func RunLoginPrompt() (Credentials, error) {
	var authMethod AuthMethod
	var token, email, apiKey string

	methodForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[AuthMethod]().
				Title("Choose authentication method").
				Options(
					huh.NewOption("API Token (Recommended)", TokenAuthMethod).
						Selected(true),
					huh.NewOption("Email + Global API Key (Legacy)", LegacyAuthMethod),
				).
				Value(&authMethod).
				Description("API Token provides secure and scoped permissions"),
		),
	).WithTheme(ui.HuhTheme())

	err := methodForm.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return Credentials{}, ErrUserCancelled
		}
		return Credentials{}, err
	}

	var credentialsForm *huh.Form

	if authMethod == TokenAuthMethod {
		credentialsForm = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("API Token").
					Description("Create tokens in: Cloudflare Dashboard → My Profile → API Tokens").
					Placeholder("Enter your API token...").
					EchoMode(huh.EchoModePassword).
					Value(&token).
					Validate(func(s string) error {
						if len(s) == 0 {
							return errors.New("token cannot be empty")
						}
						return nil
					}),
			),
		).WithTheme(ui.HuhTheme())
	} else {
		credentialsForm = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Email Address").
					Description("Find your Global API Key in: Cloudflare Dashboard → My Profile → API Tokens").
					Placeholder("your.email@example.com").
					Value(&email).
					Validate(func(s string) error {
						if len(s) == 0 {
							return errors.New("email cannot be empty")
						}
						if !isValidEmail(s) {
							return errors.New("please enter a valid email address")
						}
						return nil
					}),
				huh.NewInput().
					Title("Global API Key").
					Placeholder("Enter your Global API Key...").
					EchoMode(huh.EchoModePassword).
					Value(&apiKey).
					Validate(func(s string) error {
						if len(s) == 0 {
							return errors.New("API key cannot be empty")
						}
						return nil
					}),
			),
		).WithTheme(ui.HuhTheme())
	}

	err = credentialsForm.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return Credentials{}, ErrUserCancelled
		}
		return Credentials{}, err
	}

	creds := Credentials{
		AuthMethod: authMethod,
	}

	if authMethod == TokenAuthMethod {
		creds.Token = token
	} else {
		creds.Email = email
		creds.APIKey = apiKey
	}

	return creds, nil
}

func isValidEmail(email string) bool {
	if len(email) < 3 {
		return false
	}

	atIndex := -1
	dotIndex := -1

	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false
			}
			atIndex = i
		} else if char == '.' && atIndex != -1 {
			dotIndex = i
		}
	}

	return atIndex > 0 && dotIndex > atIndex+1 && dotIndex < len(email)-1
}
