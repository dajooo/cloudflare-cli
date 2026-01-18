package d1

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/spf13/cobra"
)

var execAccountID string

var execCmd = &cobra.Command{
	Use:   "exec <name> -- <query>",
	Short: "Execute a query against a D1 database",
	Args:  cobra.MinimumNArgs(1),
	Run: executor.NewBuilder[*cf.Client, []d1.QueryResult]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Executing query", execQueryFunc).
		Display(printExecResult).
		Build().
		CobraRun(),
}

func init() {
	execCmd.Flags().StringVar(&execAccountID, "account-id", "", "The account ID")
	D1Cmd.AddCommand(execCmd)
}

func execQueryFunc(client *cf.Client, cmd *cobra.Command, args []string, _ chan<- string) ([]d1.QueryResult, error) {
	accID, err := cloudflare.GetAccountID(client, execAccountID)
	if err != nil {
		return nil, err
	}
	dbName := args[0]

	if len(args) < 2 {
		return nil, fmt.Errorf("please provide a query after the database name, e.g. `cf d1 exec my-db -- 'SELECT * FROM users'`")
	}

	query := strings.Join(args[1:], " ")

	pager := client.D1.Database.ListAutoPaging(context.Background(), d1.DatabaseListParams{
		AccountID: cf.F(accID),
	})

	var targetUUID string
	for pager.Next() {
		db := pager.Current()
		if db.Name == dbName || db.UUID == dbName {
			targetUUID = db.UUID
			break
		}
	}
	if err := pager.Err(); err != nil {
		return nil, fmt.Errorf("error listing databases: %w", err)
	}

	if targetUUID == "" {
		return nil, fmt.Errorf("database '%s' not found", dbName)
	}

	page, err := client.D1.Database.Query(context.Background(), targetUUID, d1.DatabaseQueryParams{
		AccountID: cf.F(accID),
		Sql:       cf.F(query),
	})
	if err != nil {
		return nil, err
	}

	return page.Result, nil
}

func printExecResult(results []d1.QueryResult, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error executing query", err).Display()
		return
	}

	for i, res := range results {
		if !res.Success {
			rb.AddItem(fmt.Sprintf("Result %d (Failed)", i+1), ui.Error("Query failed"))
			continue
		}

		if len(res.Results) > 0 {
			rb.AddItem(fmt.Sprintf("Result %d", i+1), fmt.Sprintf("Rows: %d", len(res.Results)))
			// Simple dump of first result if available
			rb.AddItem("Sample", fmt.Sprintf("%v", res.Results[0]))
		} else {
			rb.AddItem(fmt.Sprintf("Result %d", i+1), "No rows returned")
		}
	}

	rb.FooterSuccess("Executed successfully %s", ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
