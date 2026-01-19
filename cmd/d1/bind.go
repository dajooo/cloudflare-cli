package d1

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/cloudflare/cloudflare-go/v6/pages"
	"github.com/spf13/cobra"
)

var boundD1ProjectKey = executor.NewKey[*pages.Project]("boundD1Project")

var bindCmd = &cobra.Command{
	Use:   "bind <database_name>",
	Short: "Bind a D1 database to a Pages project",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(boundD1ProjectKey, "Binding database").Func(bindDatabase)).
		Display(printD1BindResult).
		Run(),
}

func init() {
	bindCmd.Flags().String("to", "", "The Pages project name to bind to")
	bindCmd.Flags().String("name", "DB", "The binding name (variable name used in your code)")
	bindCmd.MarkFlagRequired("to")
	D1Cmd.AddCommand(bindCmd)
}

func bindDatabase(ctx *executor.Context, _ chan<- string) (*pages.Project, error) {
	dbName := ctx.Args[0]
	bindToProject, _ := ctx.Cmd.Flags().GetString("to")
	bindBindingName, _ := ctx.Cmd.Flags().GetString("name")

	pager := ctx.Client.D1.Database.ListAutoPaging(context.Background(), d1.DatabaseListParams{
		AccountID: cf.F(ctx.AccountID),
	})

	var dbID string
	for pager.Next() {
		db := pager.Current()
		if db.Name == dbName || db.UUID == dbName {
			dbID = db.UUID
			break
		}
	}
	if err := pager.Err(); err != nil {
		return nil, fmt.Errorf("error listing databases: %w", err)
	}
	if dbID == "" {
		return nil, fmt.Errorf("database '%s' not found", dbName)
	}

	proj, err := ctx.Client.Pages.Projects.Get(context.Background(), bindToProject, pages.ProjectGetParams{
		AccountID: cf.F(ctx.AccountID),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting project '%s': %w", bindToProject, err)
	}

	prodD1s := make(map[string]pages.ProjectDeploymentConfigsProductionD1DatabaseParam)
	if proj.DeploymentConfigs.Production.D1Databases != nil {
		for k, v := range proj.DeploymentConfigs.Production.D1Databases {
			prodD1s[k] = pages.ProjectDeploymentConfigsProductionD1DatabaseParam{
				ID: cf.F(v.ID),
			}
		}
	}
	prodD1s[bindBindingName] = pages.ProjectDeploymentConfigsProductionD1DatabaseParam{
		ID: cf.F(dbID),
	}

	prevD1s := make(map[string]pages.ProjectDeploymentConfigsPreviewD1DatabaseParam)
	if proj.DeploymentConfigs.Preview.D1Databases != nil {
		for k, v := range proj.DeploymentConfigs.Preview.D1Databases {
			prevD1s[k] = pages.ProjectDeploymentConfigsPreviewD1DatabaseParam{
				ID: cf.F(v.ID),
			}
		}
	}
	prevD1s[bindBindingName] = pages.ProjectDeploymentConfigsPreviewD1DatabaseParam{
		ID: cf.F(dbID),
	}

	return ctx.Client.Pages.Projects.Edit(context.Background(), bindToProject, pages.ProjectEditParams{
		AccountID: cf.F(ctx.AccountID),
		Project: pages.ProjectParam{
			DeploymentConfigs: cf.F(pages.ProjectDeploymentConfigsParam{
				Production: cf.F(pages.ProjectDeploymentConfigsProductionParam{
					D1Databases: cf.F(prodD1s),
				}),
				Preview: cf.F(pages.ProjectDeploymentConfigsPreviewParam{
					D1Databases: cf.F(prevD1s),
				}),
			}),
		},
	})
}

func printD1BindResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error binding database", ctx.Error).Display()
		return
	}
	proj := executor.Get(ctx, boundD1ProjectKey)
	bindBindingName, _ := ctx.Cmd.Flags().GetString("name")
	rb.FooterSuccessf("Successfully bound database to project '%s' as '%s' %s", proj.Name, bindBindingName, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
