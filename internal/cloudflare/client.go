package cloudflare

import (
	"errors"
	"fmt"
	"runtime"

	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/constants"
	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

func UserAgent() string {
	return fmt.Sprintf("cloudflare-cli/%s (%s; %s) +%s", constants.Version, runtime.GOOS, runtime.GOARCH, constants.ProjectURL)
}

func NewClient() (*cloudflare.Client, error) {
	err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	if config.Cfg.APIToken != "" {
		return cloudflare.NewClient(option.WithHeader("User-Agent", UserAgent()), option.WithAPIToken(string(config.Cfg.APIToken))), nil
	}
	if config.Cfg.APIEmail != "" && config.Cfg.APIKey != "" {
		return cloudflare.NewClient(option.WithHeader("User-Agent", UserAgent()), option.WithAPIEmail(config.Cfg.APIEmail), option.WithAPIKey(string(config.Cfg.APIKey))), nil
	}
	return nil, errors.New("you need to login first. Use `cf login` for that")
}
