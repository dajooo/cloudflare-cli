package r2

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/pages"
	"github.com/cloudflare/cloudflare-go/v6/r2"
	"github.com/spf13/cobra"
)

var bindAccountID string
var bindToProject string
var bindBindingName string

var bindCmd = &cobra.Command{
	Use:   "bind <bucket_name>",
	Short: "Bind an R2 bucket to a Pages project",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, *pages.Project]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Binding bucket", bindBucket).
		Display(printBindResult).
		Build().
		CobraRun(),
}

func init() {
	bindCmd.Flags().StringVar(&bindAccountID, "account-id", "", "The account ID")
	bindCmd.Flags().StringVar(&bindToProject, "to", "", "The Pages project name to bind to")
	bindCmd.Flags().StringVar(&bindBindingName, "name", "BUCKET", "The binding name (variable name used in your code)")
	bindCmd.MarkFlagRequired("to")
	R2Cmd.AddCommand(bindCmd)
}

func bindBucket(client *cf.Client, cmd *cobra.Command, args []string, _ chan<- string) (*pages.Project, error) {
	accID, err := cloudflare.GetAccountID(client, cmd, bindAccountID)
	if err != nil {
		return nil, err
	}
	bucketName := args[0]

	res, err := client.R2.Buckets.List(context.Background(), r2.BucketListParams{
		AccountID: cf.F(accID),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing R2 buckets: %w", err)
	}

	var found bool
	for _, b := range res.Buckets {
		if b.Name == bucketName {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("bucket '%s' not found", bucketName)
	}

	proj, err := client.Pages.Projects.Get(context.Background(), bindToProject, pages.ProjectGetParams{
		AccountID: cf.F(accID),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting project '%s': %w", bindToProject, err)
	}

	prodBuckets := make(map[string]pages.ProjectDeploymentConfigsProductionR2BucketParam)
	if proj.DeploymentConfigs.Production.R2Buckets != nil {
		for k, v := range proj.DeploymentConfigs.Production.R2Buckets {
			res := pages.ProjectDeploymentConfigsProductionR2BucketParam{
				Name: cf.F(v.Name),
			}
			if v.Jurisdiction != "" {
				res.Jurisdiction = cf.F(v.Jurisdiction)
			}
			prodBuckets[k] = res
		}
	}
	prodBuckets[bindBindingName] = pages.ProjectDeploymentConfigsProductionR2BucketParam{
		Name: cf.F(bucketName),
	}

	prevBuckets := make(map[string]pages.ProjectDeploymentConfigsPreviewR2BucketParam)
	if proj.DeploymentConfigs.Preview.R2Buckets != nil {
		for k, v := range proj.DeploymentConfigs.Preview.R2Buckets {
			res := pages.ProjectDeploymentConfigsPreviewR2BucketParam{
				Name: cf.F(v.Name),
			}
			if v.Jurisdiction != "" {
				res.Jurisdiction = cf.F(v.Jurisdiction)
			}
			prevBuckets[k] = res
		}
	}
	prevBuckets[bindBindingName] = pages.ProjectDeploymentConfigsPreviewR2BucketParam{
		Name: cf.F(bucketName),
	}

	return client.Pages.Projects.Edit(context.Background(), bindToProject, pages.ProjectEditParams{
		AccountID: cf.F(accID),
		Project: pages.ProjectParam{
			DeploymentConfigs: cf.F(pages.ProjectDeploymentConfigsParam{
				Production: cf.F(pages.ProjectDeploymentConfigsProductionParam{
					R2Buckets: cf.F(prodBuckets),
				}),
				Preview: cf.F(pages.ProjectDeploymentConfigsPreviewParam{
					R2Buckets: cf.F(prevBuckets),
				}),
			}),
		},
	})
}

func printBindResult(proj *pages.Project, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error binding R2 bucket", err).Display()
		return
	}
	rb.FooterSuccess("Successfully bound R2 bucket to project '%s' as '%s' %s", proj.Name, bindBindingName, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
