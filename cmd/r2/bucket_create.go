package r2

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/r2"
	"github.com/spf13/cobra"
)

var createdBucketKey = executor.NewKey[*r2.Bucket]("createdBucket")

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new R2 bucket",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(createdBucketKey, "Creating bucket").Func(createBucket)).
		Display(printCreateBucket).
		Run(),
}

func init() {
	bucketCmd.AddCommand(createCmd)
}

func createBucket(ctx *executor.Context, _ chan<- string) (*r2.Bucket, error) {
	return ctx.Client.R2.Buckets.New(context.Background(), r2.BucketNewParams{
		AccountID: cf.F(ctx.AccountID),
		Name:      cf.F(ctx.Args[0]),
	})
}

func printCreateBucket(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error creating bucket", ctx.Error).Display()
		return
	}
	bucket := executor.Get(ctx, createdBucketKey)
	rb.FooterSuccess("Successfully created bucket %s %s", bucket.Name, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
