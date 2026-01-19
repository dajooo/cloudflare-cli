package cmd

import (
	"context"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/flags"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/audit_logs"
	"github.com/cloudflare/cloudflare-go/v6/packages/pagination"
	"github.com/cloudflare/cloudflare-go/v6/shared"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit-logs",
	Short: "View audit logs",
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List audit logs",
	Run: executor.NewBuilder[*cf.Client, *pagination.V4PagePaginationArray[shared.AuditLog]]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching audit logs", func(client *cf.Client, cmd *cobra.Command, args []string, progress chan<- string) (*pagination.V4PagePaginationArray[shared.AuditLog], error) {
			accountID, err := cloudflare.GetAccountID(client, cmd, "")
			if err != nil {
				return nil, err
			}

			logs, err := client.AuditLogs.List(context.Background(), audit_logs.AuditLogListParams{
				AccountID: cf.F(accountID),
			})
			if err != nil {
				return nil, err
			}
			return logs, nil
		}).
		Display(func(logs *pagination.V4PagePaginationArray[shared.AuditLog], fetchDuration time.Duration, err error) {
			rb := response.New().Title("Audit Logs")
			if err != nil {
				rb.Error("Error fetching audit logs", err).Display()
				return
			}

			// logs.Result is []shared.AuditLog

			// Note: Pagination result struct has Result field
			results := logs.Result

			rb.Summary("Total (this page):", len(results))

			for i, log := range results {
				icb := response.NewItemContent().
					Add("Action:", ui.Text(log.Action.Type)).
					Add("Actor:", ui.Text(log.Actor.Email)).
					Add("When:", ui.Small(log.When.Format("2006-01-02 15:04:05")))

				if log.Resource.ID != "" {
					icb.Add("Resource:", ui.Text(log.Resource.ID))
				}

				rb.AddItem("Log #"+log.ID, icb.String())

				if i >= 10 {
					rb.FooterSuccess("Showing first 10 results")
					break
				}
			}

			if len(results) == 0 {
				rb.NoItemsMessage("No audit logs found")
			}

			rb.Display()
		}).
		Build().
		CobraRun(),
}

func init() {
	flags.RegisterAccountID(auditCmd)
	auditCmd.AddCommand(auditListCmd)
	rootCmd.AddCommand(auditCmd)
}
