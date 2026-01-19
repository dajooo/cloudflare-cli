package d1

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/spf13/cobra"
)

var queryResultsKey = executor.NewKey[[]d1.QueryResult]("queryResults")

var execCmd = &cobra.Command{
	Use:   "exec <name> -- <query>",
	Short: "Execute a query against a D1 database",
	Args:  cobra.MinimumNArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(queryResultsKey, "Executing query").Func(execQueryFunc)).
		Display(printExecResult).
		Run(),
}

func init() {
	D1Cmd.AddCommand(execCmd)
}

func execQueryFunc(ctx *executor.Context, _ chan<- string) ([]d1.QueryResult, error) {
	dbName := ctx.Args[0]

	if len(ctx.Args) < 2 {
		return nil, fmt.Errorf("please provide a query after the database name, e.g. `cf d1 exec my-db -- 'SELECT * FROM users'`")
	}

	query := strings.Join(ctx.Args[1:], " ")

	pager := ctx.Client.D1.Database.ListAutoPaging(context.Background(), d1.DatabaseListParams{
		AccountID: cf.F(ctx.AccountID),
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

	page, err := ctx.Client.D1.Database.Query(context.Background(), targetUUID, d1.DatabaseQueryParams{
		AccountID: cf.F(ctx.AccountID),
		Sql:       cf.F(query),
	})
	if err != nil {
		return nil, err
	}

	return page.Result, nil
}

func printExecResult(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error executing query", ctx.Error).Display()
		return
	}

	results := executor.Get(ctx, queryResultsKey)

	for i, res := range results {
		if !res.Success {
			rb.AddItem(fmt.Sprintf("Result %d (Failed)", i+1), ui.Error("Query failed"))
			continue
		}

		if len(res.Results) > 0 {
			firstRow, ok := res.Results[0].(map[string]interface{})
			if !ok {
				rb.AddItem(fmt.Sprintf("Result %d", i+1), fmt.Sprintf("Rows: %d", len(res.Results)))
				rb.AddItem("Sample", fmt.Sprintf("%v", res.Results[0]))
				continue
			}

			headers := make([]string, 0, len(firstRow))
			for k := range firstRow {
				headers = append(headers, k)
			}
			sort.Strings(headers)

			t := table.New().
				Border(lipgloss.HiddenBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(ui.C.Gray500)).
				Headers(headers...)

			for _, rowInterface := range res.Results {
				rowMap, ok := rowInterface.(map[string]interface{})
				if !ok {
					continue
				}
				row := make([]string, 0, len(headers))
				for _, h := range headers {
					val := rowMap[h]
					if val == nil {
						row = append(row, "NULL")
					} else {
						strVal := fmt.Sprintf("%v", val)
						strVal = strings.ReplaceAll(strVal, "\n", " ")
						strVal = strings.ReplaceAll(strVal, "\r", " ")
						strVal = strings.ReplaceAll(strVal, "\t", " ")

						if len(strVal) > 40 {
							strVal = strVal[:37] + "..."
						}
						row = append(row, strVal)
					}
				}
				t.Row(row...)
			}

			rb.AddItem(fmt.Sprintf("Result %d", i+1), t.Render())
		} else {
			rb.AddItem(fmt.Sprintf("Result %d", i+1), "No rows returned")
		}
	}

	rb.FooterSuccessf("Executed successfully %s", ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
