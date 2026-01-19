package d1

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
	"github.com/spf13/cobra"
)

var createdDatabaseKey = executor.NewKey[*d1.D1]("createdDatabase")

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new D1 database",
	Args:  cobra.ExactArgs(1),
	Run: executor.New().
		WithClient().
		WithAccountID().
		Step(executor.NewStep(createdDatabaseKey, "Creating database").Func(createDatabase)).
		Invalidates(func(ctx *executor.Context) []string {
			return []string{"d1:databases:list"}
		}).
		Display(printCreateDatabase).
		Run(),
}

func init() {
	D1Cmd.AddCommand(createCmd)
}

func createDatabase(ctx *executor.Context, _ chan<- string) (*d1.D1, error) {
	return ctx.Client.D1.Database.New(context.Background(), d1.DatabaseNewParams{
		AccountID: cf.F(ctx.AccountID),
		Name:      cf.F(ctx.Args[0]),
	})
}

func printCreateDatabase(ctx *executor.Context) {
	rb := response.New()
	if ctx.Error != nil {
		rb.Error("Error creating database", ctx.Error).Display()
		return
	}
	db := executor.Get(ctx, createdDatabaseKey)
	rb.FooterSuccessf("Successfully created database %s (%s) %s", db.Name, db.UUID, ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))).Display()
}
