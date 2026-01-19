package d1

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/spf13/cobra"
)

var databasesKey = executor.NewKey[[]d1.DatabaseListResponse]("databases")

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List D1 databases",
	Run: executor.New().
		WithClient().
		WithAccountID().
		WithPagination().
		Step(executor.NewStep(databasesKey, "Fetching databases").Func(listDatabases)).
		Display(printListDatabases).
		Run(),
}

func init() {
	pagination.RegisterFlags(listCmd)
	D1Cmd.AddCommand(listCmd)
}

func listDatabases(ctx *executor.Context, _ chan<- string) ([]d1.DatabaseListResponse, error) {
	pager := ctx.Client.D1.Database.ListAutoPaging(context.Background(), d1.DatabaseListParams{
		AccountID: cf.F(ctx.AccountID),
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

func printListDatabases(ctx *executor.Context) {
	rb := response.New().Title("D1 Databases").NoItemsMessage("No databases found")

	if ctx.Error != nil {
		rb.Error("Error listing databases", ctx.Error).Display()
		return
	}

	dbs := executor.Get(ctx, databasesKey)
	paginated, info := pagination.Paginate(dbs, ctx.Pagination)

	for _, db := range paginated {
		rb.AddItem(db.Name, ui.Muted(fmt.Sprintf("ID: %s", db.UUID)))
	}

	if len(paginated) > 0 {
		footer := info.FooterMessage("database(s)")
		footer += " " + ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))
		rb.FooterSuccess(footer)
	}

	rb.Display()
}
