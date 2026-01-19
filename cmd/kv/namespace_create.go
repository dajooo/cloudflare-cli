package kv

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/spf13/cobra"
)

var createdNamespaceKey = executor.NewKey[*kv.Namespace]("createdNamespace")

var createNamespaceCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new KV namespace",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(createdNamespaceKey, "Creating namespace").Func(createNamespace)).
		Display(printCreateNamespace).
		Run(),
}

func init() {
	namespaceCmd.AddCommand(createNamespaceCmd)
}

func createNamespace(ctx *executor.Context, _ chan<- string) (*kv.Namespace, error) {
	return ctx.Client.KV.Namespaces.New(context.Background(), kv.NamespaceNewParams{
		AccountID: cf.F(ctx.AccountID),
		Title:     cf.F(ctx.Args[0]),
	})
}

func printCreateNamespace(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error creating namespace", ctx.Error).Display()
		return
	}
	ns := executor.Get(ctx, createdNamespaceKey)
	rb.FooterSuccessf("Created namespace %s (%s) %s", ns.Title, ns.ID, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
