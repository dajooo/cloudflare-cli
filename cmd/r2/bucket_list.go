package r2

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/r2"
	"github.com/spf13/cobra"
)

var listAccountID string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List R2 buckets",
	Run: executor.NewBuilder[*cf.Client, *r2.BucketListResponse]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching buckets", listBuckets).
		Display(printListBuckets).
		Build().
		CobraRun(),
}

func init() {
	listCmd.Flags().StringVar(&listAccountID, "account-id", "", "The account ID to list buckets from")
	bucketCmd.AddCommand(listCmd)
}

func listBuckets(client *cf.Client, cmd *cobra.Command, _ []string, _ chan<- string) (*r2.BucketListResponse, error) {
	accID, err := cloudflare.GetAccountID(client, cmd, listAccountID)
	if err != nil {
		return nil, err
	}
	return client.R2.Buckets.List(context.Background(), r2.BucketListParams{
		AccountID: cf.F(accID),
	})
}

func printListBuckets(res *r2.BucketListResponse, duration time.Duration, err error) {
	rb := response.New().Title("R2 Buckets")
	if err != nil {
		rb.Error("Error listing buckets", err).Display()
		return
	}

	for _, bucket := range res.Buckets {
		rb.AddItem(bucket.Name, ui.Muted(bucket.CreationDate))
	}

	if len(res.Buckets) == 0 {
		rb.NoItemsMessage("No buckets found")
	} else {
		rb.FooterSuccess("Found %d buckets %s", len(res.Buckets), ui.Muted(fmt.Sprintf("(took %v)", duration)))
	}
	rb.Display()
}
