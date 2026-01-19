package kv

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/spf13/cobra"
)

var putKeyCmd = &cobra.Command{
	Use:   "put <key> <value>",
	Short: "Put a key-value pair into a namespace",
	Args:  cobra.ExactArgs(2),
	Run: executor.NewBuilder[*cf.Client, *kv.NamespaceValueUpdateResponse]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Putting key", putKey).
		Display(printPutKey).
		Build().
		CobraRun(),
}

func init() {
	keyCmd.AddCommand(putKeyCmd)
}

func putKey(client *cf.Client, cmd *cobra.Command, args []string, _ chan<- string) (*kv.NamespaceValueUpdateResponse, error) {
	accID, err := cloudflare.GetAccountID(client, cmd, keyAccountID)
	if err != nil {
		return nil, err
	}
	keyName := args[0]
	value := args[1]

	nsID := namespaceID
	if nsID == "" {
		nsID = config.Cfg.KVNamespaceID
	}
	if nsID == "" {
		return nil, fmt.Errorf("namespace ID is required. Use --namespace-id or 'cf kv namespace switch'")
	}

	return client.KV.Namespaces.Values.Update(context.Background(), nsID, keyName, kv.NamespaceValueUpdateParams{
		AccountID: cf.F(accID),
		Value:     cf.F(value),
	})
}

func printPutKey(res *kv.NamespaceValueUpdateResponse, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error putting key", err).Display()
		return
	}
	rb.FooterSuccess("Successfully put key %s", ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
