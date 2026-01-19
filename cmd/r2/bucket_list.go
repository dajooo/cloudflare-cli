package r2

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/r2"
	"github.com/spf13/cobra"
)

var bucketsKey = executor.NewKey[[]r2.Bucket]("buckets")

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List R2 buckets",
	Run: executor.New().
		WithClient().
		WithAccountID().
		WithPagination().
		Step(executor.NewStep(bucketsKey, "Fetching buckets").Func(listBuckets)).
		Display(printListBuckets).
		Run(),
}

func init() {
	pagination.RegisterFlags(listCmd)
	bucketCmd.AddCommand(listCmd)
}

func listBuckets(ctx *executor.Context, _ chan<- string) ([]r2.Bucket, error) {
	res, err := ctx.Client.R2.Buckets.List(context.Background(), r2.BucketListParams{
		AccountID: cf.F(ctx.AccountID),
	})
	if err != nil {
		return nil, err
	}
	return res.Buckets, nil
}

func printListBuckets(ctx *executor.Context) {
	rb := response.New().Title("R2 Buckets")

	if ctx.Error != nil {
		rb.Error("Error listing buckets", ctx.Error).Display()
		return
	}

	buckets := executor.Get(ctx, bucketsKey)
	paginated, info := pagination.Paginate(buckets, ctx.Pagination)

	for _, bucket := range paginated {
		rb.AddItem(bucket.Name, ui.Muted(bucket.CreationDate))
	}

	if len(paginated) == 0 {
		rb.NoItemsMessage("No buckets found")
	} else {
		footer := info.FooterMessage("bucket(s)")
		footer += " " + ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))
		rb.FooterSuccess(footer)
	}

	rb.Display()
}
