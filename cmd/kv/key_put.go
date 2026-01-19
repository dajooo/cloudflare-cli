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

var putResultKey = executor.NewKey[bool]("putResult")

var putKeyCmd = &cobra.Command{
	Use:   "put <key> <value>",
	Short: "Put a key-value pair into a namespace",
	Args:  cobra.ExactArgs(2),
	Run: executor.New().
		WithClient().
		WithAccountID().
		WithKVNamespace().
		Step(executor.NewStep(putResultKey, "Putting key").Func(putKey)).
		Invalidates(func(ctx *executor.Context) []string {
			return []string{"kv:namespace:" + ctx.KVNamespace + ":"}
		}).
		Display(printPutKey).
		Run(),
}

func init() {
	keyCmd.AddCommand(putKeyCmd)
}

func putKey(ctx *executor.Context, _ chan<- string) (bool, error) {
	keyName := ctx.Args[0]
	value := ctx.Args[1]

	_, err := ctx.Client.KV.Namespaces.Values.Update(context.Background(), ctx.KVNamespace, keyName, kv.NamespaceValueUpdateParams{
		AccountID: cf.F(ctx.AccountID),
		Value:     cf.F(value),
	})
	return err == nil, err
}

func printPutKey(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error putting key", ctx.Error).Display()
		return
	}
	rb.FooterSuccessf("Successfully put key %s", ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
