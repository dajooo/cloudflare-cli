package d1

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/spf13/cobra"
)

var listAccountID string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List D1 databases",
	Run: executor.NewBuilder[*cf.Client, []d1.DatabaseListResponse]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching databases", listDatabases).
		Display(printListDatabases).
		Build().
		CobraRun(),
}

func init() {
	listCmd.Flags().StringVar(&listAccountID, "account-id", "", "The account ID to list databases from")
	D1Cmd.AddCommand(listCmd)
}

func listDatabases(client *cf.Client, _ *cobra.Command, _ []string, _ chan<- string) ([]d1.DatabaseListResponse, error) {
	accID, err := cloudflare.GetAccountID(client, listAccountID)
	if err != nil {
		return nil, err
	}
	// Using AutoPaging if possible
	pager := client.D1.Database.ListAutoPaging(context.Background(), d1.DatabaseListParams{
		AccountID: cf.F(accID),
	})

	var all []d1.DatabaseListResponse
	for pager.Next() {
		all = append(all, pager.Current())
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}
	return all, nil
}

func printListDatabases(dbs []d1.DatabaseListResponse, duration time.Duration, err error) {
	rb := response.New().Title("D1 Databases").NoItemsMessage("No databases found")
	if err != nil {
		rb.Error("Error listing databases", err).Display()
		return
	}

	for _, db := range dbs {
		rb.AddItem(db.Name, ui.Muted(db.UUID))
	}

	if len(dbs) == 0 {
		rb.FooterSuccess("Found %d databases %s", len(dbs), ui.Muted(fmt.Sprintf("(took %v)", duration)))
	}
	rb.Display()
}
