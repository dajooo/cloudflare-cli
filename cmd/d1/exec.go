package d1

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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
	// The account ID is now fetched in the previous step and passed as the first arg
	accID := args[0]
	dbName := args[1]

	if len(args) < 3 {
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
			// Try to cast the first row to a map
			firstRow, ok := res.Results[0].(map[string]interface{})
			if !ok {
				// Fallback to simple dump if not a map
				rb.AddItem(fmt.Sprintf("Result %d", i+1), fmt.Sprintf("Rows: %d", len(res.Results)))
				rb.AddItem("Sample", fmt.Sprintf("%v", res.Results[0]))
				continue
			}

			// Get headers from the first row of map keys
			headers := make([]string, 0, len(firstRow))
			for k := range firstRow {
				headers = append(headers, k)
			}
			sort.Strings(headers)

			// Build table
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
						// Remove newlines and tabs to keep the table clean
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

	rb.FooterSuccess("Executed successfully %s", ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
