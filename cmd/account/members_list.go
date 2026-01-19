package account

import (
	"context"
	"fmt"
	"time"

	"dario.lol/cf/internal/cloudflare"
	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/flags"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/spf13/cobra"
)

var membersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage account members",
}

var membersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List members of the current account",
	Run: executor.NewBuilder[*cf.Client, []accounts.Member]().
		Setup("Decrypting configuration", cloudflare.NewClient).
		Fetch("Fetching members", func(client *cf.Client, cmd *cobra.Command, args []string, progress chan<- string) ([]accounts.Member, error) {
			accountID, err := cloudflare.GetAccountID(client, cmd, "")
			if err != nil {
				return nil, err
			}

			membersList, err := client.Accounts.Members.List(context.Background(), accounts.MemberListParams{
				AccountID: cf.F(accountID),
			})
			if err != nil {
				return nil, err
			}
			return membersList.Result, nil
		}).
		Display(func(members []accounts.Member, fetchDuration time.Duration, err error) {
			rb := response.New().Title("Account Members")
			if err != nil {
				rb.Error("Error fetching members", err).Display()
				return
			}

			rb.Summary("Total:", len(members))

			for _, member := range members {
				roles := []string{}
				for _, role := range member.Roles {
					roles = append(roles, role.Name)
				}

				statusColor := ui.Success
				if string(member.Status) != "accepted" {
					statusColor = ui.Warning
				}

				icb := response.NewItemContent().
					Add("Email:", ui.Text(member.User.Email)).
					Add("Status:", statusColor(string(member.Status))).
					Add("Roles:", ui.Muted(fmt.Sprintf("%v", roles)))

				rb.AddItem(member.User.FirstName+" "+member.User.LastName, icb.String())
			}

			if len(members) > 0 {
				rb.FooterSuccess("Found %d member(s)", len(members))
			} else {
				rb.NoItemsMessage("No members found")
			}

			rb.Display()
		}).
		Build().
		CobraRun(),
}

func init() {
	flags.RegisterAccountID(membersCmd)
	membersCmd.AddCommand(membersListCmd)
	AccountCmd.AddCommand(membersCmd)
}
