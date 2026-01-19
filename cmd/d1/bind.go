package d1

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/cloudflare/cloudflare-go/v6/pages"
	"github.com/spf13/cobra"
)

var bindAccountID string
var bindToProject string
var bindBindingName string

var bindCmd = &cobra.Command{
	Use:   "bind <database_name>",
	Short: "Bind a D1 database to a Pages project",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, *pages.Project]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Binding database", bindDatabase).
		Display(printBindResult).
		Build().
		CobraRun(),
}

func init() {
	bindCmd.Flags().StringVar(&bindAccountID, "account-id", "", "The account ID")
	bindCmd.Flags().StringVar(&bindToProject, "to", "", "The Pages project name to bind to")
	bindCmd.Flags().StringVar(&bindBindingName, "name", "DB", "The binding name (variable name used in your code)")
	bindCmd.MarkFlagRequired("to")
	D1Cmd.AddCommand(bindCmd)
}

func bindDatabase(client *cf.Client, cmd *cobra.Command, args []string, _ chan<- string) (*pages.Project, error) {
	accID, err := cloudflare.GetAccountID(client, cmd, bindAccountID)
	if err != nil {
		return nil, err
	}
	dbName := args[0]

	pager := client.D1.Database.ListAutoPaging(context.Background(), d1.DatabaseListParams{
		AccountID: cf.F(accID),
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

	proj, err := client.Pages.Projects.Get(context.Background(), bindToProject, pages.ProjectGetParams{
		AccountID: cf.F(accID),
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

	return client.Pages.Projects.Edit(context.Background(), bindToProject, pages.ProjectEditParams{
		AccountID: cf.F(accID),
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

func printBindResult(proj *pages.Project, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error binding database", err).Display()
		return
	}
	rb.FooterSuccess("Successfully bound database to project '%s' as '%s' %s", proj.Name, bindBindingName, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
