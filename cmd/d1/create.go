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
	"github.com/spf13/cobra"
)

var createAccountID string

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new D1 database",
	Args:  cobra.ExactArgs(1),
	Run: executor.NewBuilder[*cf.Client, *d1.D1]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Creating database", createDatabase).
		Display(printCreateDatabase).
		Build().
		CobraRun(),
}

func init() {
	createCmd.Flags().StringVar(&createAccountID, "account-id", "", "The account ID to create the database in")
	D1Cmd.AddCommand(createCmd)
}

func createDatabase(client *cf.Client, _ *cobra.Command, args []string, _ chan<- string) (*d1.D1, error) {
	accID, err := cloudflare.GetAccountID(client, createAccountID)
	if err != nil {
		return nil, err
	}
	return client.D1.Database.New(context.Background(), d1.DatabaseNewParams{
		AccountID: cf.F(accID),
		Name:      cf.F(args[0]),
	})
}

func printCreateDatabase(db *d1.D1, duration time.Duration, err error) {
	rb := response.New()
	if err != nil {
		rb.Error("Error creating database", err).Display()
		return
	}
	rb.FooterSuccess("Successfully created database %s (%s) %s", db.Name, db.UUID, ui.Muted(fmt.Sprintf("(took %v)", duration))).Display()
}
