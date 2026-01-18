package kv

import (
	"context"
	"fmt"
	"io"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/spf13/cobra"
)

var keyAccountID string
var namespaceID string

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage KV keys",
}

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
	keyCmd.PersistentFlags().StringVar(&keyAccountID, "account-id", "", "The account ID")
	keyCmd.PersistentFlags().StringVar(&namespaceID, "namespace-id", "", "The namespace ID")

	putKeyCmd.MarkFlagRequired("namespace-id")
	getKeyCmd.MarkFlagRequired("namespace-id")

	keyCmd.AddCommand(putKeyCmd)
	keyCmd.AddCommand(getKeyCmd)
	KVCmd.AddCommand(keyCmd)
}

func putKey(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*kv.NamespaceValueUpdateResponse, error) {
	accID, err := cloudflare.GetAccountID(client, keyAccountID)
	if err != nil {
		return nil, err
	}
	keyName := args[0]
	value := args[1]

	return client.KV.Namespaces.Values.Update(context.Background(), namespaceID, keyName, kv.NamespaceValueUpdateParams{
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

func getKey(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) ([]byte, error) {
	accID, err := cloudflare.GetAccountID(client, keyAccountID)
	if err != nil {
		return nil, err
	}
	keyName := args[0]
	resp, err := client.KV.Namespaces.Values.Get(context.Background(), namespaceID, keyName, kv.NamespaceValueGetParams{
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
