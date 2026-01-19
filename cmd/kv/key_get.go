package kv

import (
	"context"
	"fmt"
	"io"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/spf13/cobra"
)

var getKeyCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a value from a namespace",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, []byte]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Getting key", getKey).
		Display(printGetKey).
		Build().
		CobraRun(),
}

func init() {
	keyCmd.AddCommand(getKeyCmd)
}

func getKey(client *cf.Client, cmd *cobra.Command, args []string, _ chan<- string) ([]byte, error) {
	accID, err := cloudflare.GetAccountID(client, cmd, keyAccountID)
	if err != nil {
		return nil, err
	}
	keyName := args[0]

	nsID := namespaceID
	if nsID == "" {
		nsID = config.Cfg.KVNamespaceID
	}
	if nsID == "" {
		return nil, fmt.Errorf("namespace ID is required. Use --namespace-id or 'cf kv namespace switch'")
	}

	resp, err := client.KV.Namespaces.Values.Get(context.Background(), nsID, keyName, kv.NamespaceValueGetParams{
		AccountID: cf.F(accID),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func printGetKey(val []byte, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error getting key", err).Display()
		return
	}
	rb.AddItem("Value", string(val))
	rb.Display()
}
