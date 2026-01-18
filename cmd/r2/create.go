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

var accountID string

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new R2 bucket",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, *r2.Bucket]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Creating bucket", createBucket).
		Display(printCreateBucket).
		Build().
		CobraRun(),
}

func init() {
	createCmd.Flags().StringVar(&accountID, "account-id", "", "The account ID to create the bucket in")
	bucketCmd.AddCommand(createCmd)
}

func createBucket(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*r2.Bucket, error) {
	accID, err := cloudflare.GetAccountID(client, accountID)
	if err != nil {
		return nil, err
	}
	return client.R2.Buckets.New(context.Background(), r2.BucketNewParams{
		AccountID: cf.F(accID),
		Name:      cf.F(args[0]),
	})
}

func printCreateBucket(bucket *r2.Bucket, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error creating bucket", err).Display()
		return
	}
	rb.FooterSuccess("Successfully created bucket %s %s", bucket.Name, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
