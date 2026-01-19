package kv

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/spf13/cobra"
)

var namespacesKey = executor.NewKey[[]kv.Namespace]("namespaces")

var listNamespaceCmd = &cobra.Command{
	Use:   "list",
	Short: "List KV namespaces",
	Run: executor.New().
		WithClient().
		WithAccountID().
		WithPagination().
		Step(executor.NewStep(namespacesKey, "Listing namespaces").Func(listNamespaces)).
		Display(printListNamespaces).
		Run(),
}

func init() {
	pagination.RegisterFlags(listNamespaceCmd)
	namespaceCmd.AddCommand(listNamespaceCmd)
}

func listNamespaces(ctx *executor.Context, _ chan<- string) ([]kv.Namespace, error) {
	pager := ctx.Client.KV.Namespaces.ListAutoPaging(context.Background(), kv.NamespaceListParams{
		AccountID: cf.F(ctx.AccountID),
	})
	var all []kv.Namespace
	for pager.Next() {
		all = append(all, pager.Current())
	}
	return all, pager.Err()
}

func printListNamespaces(ctx *executor.Context) {
	rb := response.New().Title("KV Namespaces")

	if ctx.Error != nil {
		rb.Error("Error listing namespaces", ctx.Error).Display()
		return
	}

	nss := executor.Get(ctx, namespacesKey)
	paginated, info := pagination.Paginate(nss, ctx.Pagination)

	for _, ns := range paginated {
		rb.AddItem(ns.Title, ui.Muted(ns.ID))
	}

	if len(paginated) == 0 {
		rb.NoItemsMessage("No namespaces found")
	} else {
		footer := fmt.Sprintf("Showing %d of %d namespace(s)", info.Showing, info.Total)
		footer += " " + ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))
		rb.FooterSuccess(footer)
	}

	rb.Display()
}
