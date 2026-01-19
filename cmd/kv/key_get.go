package kv

import (
	"context"
	"fmt"
	"io"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/spf13/cobra"
)

var keyValueKey = executor.NewKey[[]byte]("keyValue")

var getKeyCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a value from a namespace",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		WithKVNamespace().
		Step(executor.NewStep(keyValueKey, "Getting key").Func(getKey)).
		Display(printGetKey).
		Run(),
}

func init() {
	keyCmd.AddCommand(getKeyCmd)
}

func getKey(ctx *executor.Context, _ chan<- string) ([]byte, error) {
	keyName := ctx.Args[0]

	resp, err := ctx.Client.KV.Namespaces.Values.Get(context.Background(), ctx.KVNamespace, keyName, kv.NamespaceValueGetParams{
		AccountID: cf.F(ctx.AccountID),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func printGetKey(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error getting key", ctx.Error).Display()
		return
	}
	val := executor.Get(ctx, keyValueKey)
	fmt.Println(string(val))
}
