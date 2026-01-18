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

var namespaceAccountID string

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Manage KV namespaces",
}

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

var listNamespaceCmd = &cobra.Command{
	Use:   "list",
	Short: "List KV namespaces",
	Run: executor.NewBuilder[*cf.Client, []kv.Namespace]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Listing namespaces", listNamespaces).
		Display(printListNamespaces).
		Build().
		CobraRun(),
}

func init() {
	namespaceCmd.PersistentFlags().StringVar(&namespaceAccountID, "account-id", "", "The account ID")

	namespaceCmd.AddCommand(createNamespaceCmd)
	namespaceCmd.AddCommand(listNamespaceCmd)
	KVCmd.AddCommand(namespaceCmd)
}

func createNamespace(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*kv.Namespace, error) {
	accID, err := cloudflare.GetAccountID(client, namespaceAccountID)
	if err != nil {
		return nil, err
	}
	return client.KV.Namespaces.New(context.Background(), kv.NamespaceNewParams{
		AccountID: cf.F(accID),
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

func listNamespaces(client *cf.Client, _ *cobra.Command, _ []string, _ chan<- string) ([]kv.Namespace, error) {
	accID, err := cloudflare.GetAccountID(client, namespaceAccountID)
	if err != nil {
		return nil, err
	}
	pager := client.KV.Namespaces.ListAutoPaging(context.Background(), kv.NamespaceListParams{
		AccountID: cf.F(accID),
	})
	var all []kv.Namespace
	for pager.Next() {
		all = append(all, pager.Current())
	}
	return all, pager.Err()
}

func printListNamespaces(nss []kv.Namespace, duration time.Duration, err error) {
	rb := response.New().Title("KV Namespaces")
	if err != nil {
		rb.Error("Error listing namespaces", err).Display()
		return
	}
	for _, ns := range nss {
		rb.AddItem(ns.Title, ui.Muted(ns.ID))
	}
	if len(nss) == 0 {
		rb.NoItemsMessage("No namespaces found")
	} else {
		rb.FooterSuccess("Found %d namespaces %s", len(nss), ui.Muted(fmt.Sprintf("(took %v)", duration)))
	}
	rb.Display()
}
