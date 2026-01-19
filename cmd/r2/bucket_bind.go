package r2

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/pages"
	"github.com/cloudflare/cloudflare-go/v6/r2"
	"github.com/spf13/cobra"
)

var boundR2ProjectKey = executor.NewKey[*pages.Project]("boundR2Project")

var bindCmd = &cobra.Command{
	Use:   "bind <bucket_name>",
	Short: "Bind an R2 bucket to a Pages project",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(boundR2ProjectKey, "Binding bucket").Func(bindBucket)).
		Display(printR2BindResult).
		Run(),
}

func init() {
	bindCmd.Flags().String("to", "", "The Pages project name to bind to")
	bindCmd.Flags().String("name", "BUCKET", "The binding name (variable name used in your code)")
	bindCmd.MarkFlagRequired("to")
	R2Cmd.AddCommand(bindCmd)
}

func bindBucket(ctx *executor.Context, _ chan<- string) (*pages.Project, error) {
	bucketName := ctx.Args[0]
	bindToProject, _ := ctx.Cmd.Flags().GetString("to")
	bindBindingName, _ := ctx.Cmd.Flags().GetString("name")

	res, err := ctx.Client.R2.Buckets.List(context.Background(), r2.BucketListParams{
		AccountID: cf.F(ctx.AccountID),
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

	proj, err := ctx.Client.Pages.Projects.Get(context.Background(), bindToProject, pages.ProjectGetParams{
		AccountID: cf.F(ctx.AccountID),
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

	return ctx.Client.Pages.Projects.Edit(context.Background(), bindToProject, pages.ProjectEditParams{
		AccountID: cf.F(ctx.AccountID),
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

func printR2BindResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error binding R2 bucket", ctx.Error).Display()
		return
	}
	proj := executor.Get(ctx, boundR2ProjectKey)
	bindBindingName, _ := ctx.Cmd.Flags().GetString("name")
	rb.FooterSuccess("Successfully bound R2 bucket to project '%s' as '%s' %s", proj.Name, bindBindingName, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
