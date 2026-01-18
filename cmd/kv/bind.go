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
	"github.com/cloudflare/cloudflare-go/v6/pages"
	"github.com/spf13/cobra"
)

var bindAccountID string
var bindToProject string
var bindBindingName string

var bindCmd = &cobra.Command{
	Use:   "bind <namespace_name_or_id>",
	Short: "Bind a KV namespace to a Pages project",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, *pages.Project]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Binding namespace", bindNamespace).
		Display(printBindResult).
		Build().
		CobraRun(),
}

func init() {
	bindCmd.Flags().StringVar(&bindAccountID, "account-id", "", "The account ID")
	bindCmd.Flags().StringVar(&bindToProject, "to", "", "The Pages project name to bind to")
	bindCmd.Flags().StringVar(&bindBindingName, "name", "KV", "The binding name (variable name used in your code)")
	bindCmd.MarkFlagRequired("to")
	KVCmd.AddCommand(bindCmd)
}

func bindNamespace(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*pages.Project, error) {
	accID, err := cloudflare.GetAccountID(client, bindAccountID)
	if err != nil {
		return nil, err
	}
	nsNameOrID := args[0]

	// 1. Resolve KV ID
	pager := client.KV.Namespaces.ListAutoPaging(context.Background(), kv.NamespaceListParams{
		AccountID: cf.F(accID),
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
		// Assuming user provided an ID directly if not found by name?
		// But best validation is to assume it might be an ID if format matches, mostly it's safer to resolve.
		// If not found in list, and looks like an ID, maybe?
		// For now, strict resolution.
		return nil, fmt.Errorf("KV namespace '%s' not found", nsNameOrID)
	}

	// 2. Refresh Project
	proj, err := client.Pages.Projects.Get(context.Background(), bindToProject, pages.ProjectGetParams{
		AccountID: cf.F(accID),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting project '%s': %w", bindToProject, err)
	}

	// 3. Prepare Update

	// Production
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

	// Preview
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

	// 4. Update Project
	return client.Pages.Projects.Edit(context.Background(), bindToProject, pages.ProjectEditParams{
		AccountID: cf.F(accID),
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

func printBindResult(proj *pages.Project, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error binding KV namespace", err).Display()
		return
	}
	rb.FooterSuccess("Successfully bound KV namespace to project '%s' as '%s' %s", proj.Name, bindBindingName, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
