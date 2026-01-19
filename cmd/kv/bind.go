package kv

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/cloudflare/cloudflare-go/v6/pages"
	"github.com/spf13/cobra"
)

var boundKVProjectKey = executor.NewKey[*pages.Project]("boundKVProject")

var bindCmd = &cobra.Command{
	Use:   "bind <namespace_name_or_id>",
	Short: "Bind a KV namespace to a Pages project",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(boundKVProjectKey, "Binding namespace").Func(bindNamespace)).
		Display(printKVBindResult).
		Run(),
}

func init() {
	bindCmd.Flags().String("to", "", "The Pages project name to bind to")
	bindCmd.Flags().String("name", "KV", "The binding name (variable name used in your code)")
	bindCmd.MarkFlagRequired("to")
	KVCmd.AddCommand(bindCmd)
}

func bindNamespace(ctx *executor.Context, _ chan<- string) (*pages.Project, error) {
	nsNameOrID := ctx.Args[0]
	bindToProject, _ := ctx.Cmd.Flags().GetString("to")
	bindBindingName, _ := ctx.Cmd.Flags().GetString("name")

	pager := ctx.Client.KV.Namespaces.ListAutoPaging(context.Background(), kv.NamespaceListParams{
		AccountID: cf.F(ctx.AccountID),
	})

	var nsID string
	for pager.Next() {
		ns := pager.Current()
		if ns.Title == nsNameOrID || ns.ID == nsNameOrID {
			nsID = ns.ID
			break
		}
	}
	if err := pager.Err(); err != nil {
		return nil, fmt.Errorf("error listing KV namespaces: %w", err)
	}
	if nsID == "" {
		return nil, fmt.Errorf("KV namespace '%s' not found", nsNameOrID)
	}

	proj, err := ctx.Client.Pages.Projects.Get(context.Background(), bindToProject, pages.ProjectGetParams{
		AccountID: cf.F(ctx.AccountID),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting project '%s': %w", bindToProject, err)
	}

	prodKVs := make(map[string]pages.ProjectDeploymentConfigsProductionKVNamespaceParam)
	if proj.DeploymentConfigs.Production.KVNamespaces != nil {
		for k, v := range proj.DeploymentConfigs.Production.KVNamespaces {
			prodKVs[k] = pages.ProjectDeploymentConfigsProductionKVNamespaceParam{
				NamespaceID: cf.F(v.NamespaceID),
			}
		}
	}
	prodKVs[bindBindingName] = pages.ProjectDeploymentConfigsProductionKVNamespaceParam{
		NamespaceID: cf.F(nsID),
	}

	prevKVs := make(map[string]pages.ProjectDeploymentConfigsPreviewKVNamespaceParam)
	if proj.DeploymentConfigs.Preview.KVNamespaces != nil {
		for k, v := range proj.DeploymentConfigs.Preview.KVNamespaces {
			prevKVs[k] = pages.ProjectDeploymentConfigsPreviewKVNamespaceParam{
				NamespaceID: cf.F(v.NamespaceID),
			}
		}
	}
	prevKVs[bindBindingName] = pages.ProjectDeploymentConfigsPreviewKVNamespaceParam{
		NamespaceID: cf.F(nsID),
	}

	return ctx.Client.Pages.Projects.Edit(context.Background(), bindToProject, pages.ProjectEditParams{
		AccountID: cf.F(ctx.AccountID),
		Project: pages.ProjectParam{
			DeploymentConfigs: cf.F(pages.ProjectDeploymentConfigsParam{
				Production: cf.F(pages.ProjectDeploymentConfigsProductionParam{
					KVNamespaces: cf.F(prodKVs),
				}),
				Preview: cf.F(pages.ProjectDeploymentConfigsPreviewParam{
					KVNamespaces: cf.F(prevKVs),
				}),
			}),
		},
	})
}

func printKVBindResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error binding KV namespace", ctx.Error).Display()
		return
	}
	proj := executor.Get(ctx, boundKVProjectKey)
	bindBindingName, _ := ctx.Cmd.Flags().GetString("name")
	rb.FooterSuccessf("Successfully bound KV namespace to project '%s' as '%s' %s", proj.Name, bindBindingName, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
