package cmd

import (
	"context"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/cache"
	"github.com/dajooo/cloudflare-cli/internal/cloudflare"
	"github.com/dajooo/cloudflare-cli/internal/ui"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage Cloudflare cache",
}

var cachePurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges the Cloudflare cache",
	Run:   executeCachePurge,
}

var zoneName string
var all bool
var files []string
var tags []string

func init() {
	cachePurgeCmd.Flags().StringVar(&zoneName, "zone", "", "The zone to purge the cache for")
	cachePurgeCmd.Flags().BoolVar(&all, "all", false, "Purge all files")
	cachePurgeCmd.Flags().StringSliceVar(&files, "files", []string{}, "A list of files to purge")
	cachePurgeCmd.Flags().StringSliceVar(&tags, "tags", []string{}, "A list of tags to purge")
	cacheCmd.AddCommand(cachePurgeCmd)
	rootCmd.AddCommand(cacheCmd)
}

func executeCachePurge(cmd *cobra.Command, args []string) {
	client, err := cloudflare.NewClient()
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error loading config", err))
		return
	}

	zoneID, err := cloudflare.ZoneIDByName(client, zoneName)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error finding zone", err))
		return
	}

	var body cache.CachePurgeParamsBodyUnion
	if all {
		body = &cache.CachePurgeParamsBodyCachePurgeEverything{}
	} else if len(files) > 0 {
		body = &cache.CachePurgeParamsBodyCachePurgeSingleFile{
			Files: cf.F(files),
		}
	} else if len(tags) > 0 {
		body = &cache.CachePurgeParamsBodyCachePurgeFlexPurgeByTags{
			Tags: cf.F(tags),
		}
	} else {
		fmt.Println(ui.Warning("Please specify what to purge"))
		return
	}

	params := cache.CachePurgeParams{
		ZoneID: cf.F(zoneID),
		Body:   body,
	}

	_, err = client.Cache.Purge(context.Background(), params)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Error purging cache", err))
		return
	}

	fmt.Println(ui.Success("Successfully purged cache"))
}
