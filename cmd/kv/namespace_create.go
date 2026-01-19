package kv

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/spf13/cobra"
)

var createNamespaceCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new KV namespace",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, *kv.Namespace]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Creating namespace", createNamespace).
		Display(printCreateNamespace).
		Build().
		CobraRun(),
}

func init() {
	namespaceCmd.AddCommand(createNamespaceCmd)
}

func createNamespace(client *cf.Client, cmd *cobra.Command, args []string, _ chan<- string) (*kv.Namespace, error) {
	accountID, err := cloudflare.GetAccountID(client, cmd, namespaceAccountID)
	if err != nil {
		return nil, err
	}
	return client.KV.Namespaces.New(context.Background(), kv.NamespaceNewParams{
		AccountID: cf.F(accountID),
		Title:     cf.F(args[0]),
	})
}

func printCreateNamespace(ns *kv.Namespace, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error creating namespace", err).Display()
		return
	}
	rb.FooterSuccess("Created namespace %s (%s) %s", ns.Title, ns.ID, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
