package kv

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/config"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/spf13/cobra"
)

var switchedNamespaceKey = executor.NewKey[kv.Namespace]("switchedNamespace")

var switchNamespaceCmd = &cobra.Command{
	Use:   "switch <name-or-id>",
	Short: "Switch the active KV namespace context",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(switchedNamespaceKey, "Verifying namespace").Func(runNamespaceSwitch)).
		Display(printNamespaceSwitch).
		Run(),
}

func init() {
	namespaceCmd.AddCommand(switchNamespaceCmd)
}

func runNamespaceSwitch(ctx *executor.Context, _ chan<- string) (kv.Namespace, error) {
	input := ctx.Args[0]
	var selectedNamespace kv.Namespace

	pager := ctx.Client.KV.Namespaces.ListAutoPaging(context.Background(), kv.NamespaceListParams{
		AccountID: cf.F(ctx.AccountID),
	})

	var matches []kv.Namespace
	for pager.Next() {
		ns := pager.Current()
		if ns.ID == input || ns.Title == input {
			matches = append(matches, ns)
		}
	}
	if err := pager.Err(); err != nil {
		return kv.Namespace{}, fmt.Errorf("failed to list namespaces: %w", err)
	}

	if len(matches) == 1 {
		selectedNamespace = matches[0]
	} else if len(matches) > 1 {
		msg := fmt.Sprintf("Multiple namespaces found matching '%s':", input)
		for _, ns := range matches {
			msg += fmt.Sprintf("\n - %s (%s)", ns.Title, ns.ID)
		}
		return kv.Namespace{}, fmt.Errorf("%s\nPlease specify the Namespace ID.", msg)
	} else {
		return kv.Namespace{}, fmt.Errorf("namespace not found with title or ID '%s'", input)
	}

	config.Cfg.KVNamespaceID = selectedNamespace.ID
	if err := config.SaveConfig(); err != nil {
		return kv.Namespace{}, fmt.Errorf("failed to save config: %w", err)
	}

	return selectedNamespace, nil
}

func printNamespaceSwitch(ctx *executor.Context) {
	if ctx.Error != nil {
		response.New().Error("Failed to switch namespace", ctx.Error).Display()
		return
	}
	ns := executor.Get(ctx, switchedNamespaceKey)
	response.New().
		Title("Namespace Switched").
		FooterSuccess("Switched context to namespace %s (%s)", ui.Text(ns.Title), ui.Muted(ns.ID)).
		Display()
}
