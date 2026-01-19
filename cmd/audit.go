package cmd

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/flags"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/audit_logs"
	"github.com/cloudflare/cloudflare-go/v6/shared"
	"github.com/spf13/cobra"
)

var auditLogsKey = executor.NewKey[[]shared.AuditLog]("auditLogs")

var auditCmd = &cobra.Command{
	Use:   "audit-logs",
	Short: "View audit logs",
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List audit logs",
	Run: executor.New().
		WithClient().
		WithAccountID().
		WithPagination().
		Step(executor.NewStep(auditLogsKey, "Fetching audit logs").Func(fetchAuditLogs)).
		Display(printAuditLogs).
		Run(),
}

func init() {
	flags.RegisterAccountID(auditCmd)
	pagination.RegisterFlags(auditListCmd)
	auditCmd.AddCommand(auditListCmd)
	rootCmd.AddCommand(auditCmd)
}

func fetchAuditLogs(ctx *executor.Context, _ chan<- string) ([]shared.AuditLog, error) {
	logs, err := ctx.Client.AuditLogs.List(context.Background(), audit_logs.AuditLogListParams{
		AccountID: cf.F(ctx.AccountID),
	})
	if err != nil {
		return nil, err
	}
	return logs.Result, nil
}

func printAuditLogs(ctx *executor.Context) {
	rb := response.New().Title("Audit Logs")

	if ctx.Error != nil {
		rb.Error("Error fetching audit logs", ctx.Error).Display()
		return
	}

	logs := executor.Get(ctx, auditLogsKey)
	paginated, info := pagination.Paginate(logs, ctx.Pagination)

	rb.Summary("Total:", info.Total)

	for _, log := range paginated {
		icb := response.NewItemContent().
			Add("Action:", ui.Text(log.Action.Type)).
			Add("Actor:", ui.Text(log.Actor.Email)).
			Add("When:", ui.Small(log.When.Format("2006-01-02 15:04:05")))

		if log.Resource.ID != "" {
			icb.Add("Resource:", ui.Text(log.Resource.ID))
		}

		rb.AddItem("Log #"+log.ID, icb.String())
	}

	if len(paginated) == 0 {
		rb.NoItemsMessage("No audit logs found")
	} else {
		footer := fmt.Sprintf("Showing %d of %d log(s)", info.Showing, info.Total)
		footer += " " + ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))
		rb.FooterSuccess(footer)
	}

	rb.Display()
}
