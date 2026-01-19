package cmd

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/cache"
	"github.com/spf13/cobra"
)

var purgeResultKey = executor.NewKey[bool]("purgeResult")

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage Cloudflare cache",
}

var cachePurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges the Cloudflare cache",
	Run: executor.New().
		WithClient().
		Step(executor.NewStep(purgeResultKey, "Purging cache").Func(purgeCache)).
		Display(printPurgeResult).
		Run(),
}

func init() {
	cachePurgeCmd.Flags().String("zone", "", "The zone to purge the cache for")
	cachePurgeCmd.Flags().Bool("all", false, "Purge all files")
	cachePurgeCmd.Flags().StringSlice("files", []string{}, "A list of files to purge")
	cachePurgeCmd.Flags().StringSlice("tags", []string{}, "A list of tags to purge")
	cacheCmd.AddCommand(cachePurgeCmd)
	rootCmd.AddCommand(cacheCmd)
}

func purgeCache(ctx *executor.Context, _ chan<- string) (bool, error) {
	zoneIdentifier, _ := ctx.Cmd.Flags().GetString("zone")
	all, _ := ctx.Cmd.Flags().GetBool("all")
	files, _ := ctx.Cmd.Flags().GetStringSlice("files")
	tags, _ := ctx.Cmd.Flags().GetStringSlice("tags")

	if zoneIdentifier == "" {
		return false, fmt.Errorf("the --zone flag is required")
	}
	zoneID, _, err := cloudflare.LookupZone(ctx.Client, zoneIdentifier)
	if err != nil {
		return false, fmt.Errorf("error finding zone: %w", err)
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
		return false, fmt.Errorf("please specify what to purge with --all, --files, or --tags")
	}

	params := cache.CachePurgeParams{
		ZoneID: cf.F(zoneID),
		Body:   body,
	}

	_, err = ctx.Client.Cache.Purge(context.Background(), params)
	return err == nil, err
}

func printPurgeResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error purging cache", ctx.Error).Display()
		return
	}
	rb.FooterSuccess("Successfully purged cache %s", ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
