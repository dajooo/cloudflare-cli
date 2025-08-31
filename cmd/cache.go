package cmd

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage Cloudflare cache",
}

var cachePurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges the Cloudflare cache",
	Run: executor.NewBuilder[*cf.Client, any]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Purging cache", purgeCache).
		Display(printPurgeResult).
		Build().
		CobraRun(),
}

var zoneIdentifier string
var all bool
var files []string
var tags []string

func init() {
	cachePurgeCmd.Flags().StringVar(&zoneIdentifier, "zone", "", "The zone to purge the cache for")
	cachePurgeCmd.Flags().BoolVar(&all, "all", false, "Purge all files")
	cachePurgeCmd.Flags().StringSliceVar(&files, "files", []string{}, "A list of files to purge")
	cachePurgeCmd.Flags().StringSliceVar(&tags, "tags", []string{}, "A list of tags to purge")
	cacheCmd.AddCommand(cachePurgeCmd)
	rootCmd.AddCommand(cacheCmd)
}

func purgeCache(client *cf.Client, _ *cobra.Command, _ []string, _ chan<- string) (any, error) {
	if zoneIdentifier == "" {
		return nil, fmt.Errorf("the --zone flag is required")
	}
	zoneID, _, err := cloudflare.LookupZone(client, zoneIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error finding zone: %w", err)
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
		return nil, fmt.Errorf("please specify what to purge with --all, --files, or --tags")
	}

	params := cache.CachePurgeParams{
		ZoneID: cf.F(zoneID),
		Body:   body,
	}

	_, err = client.Cache.Purge(context.Background(), params)
	return nil, err
}

func printPurgeResult(_ any, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error purging cache", err).Display()
		return
	}
	rb.FooterSuccess("Successfully purged cache %s", ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
