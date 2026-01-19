package account

import (
	"context"
	"fmt"

	"dario.lol/cf/internal/executor"
	"dario.lol/cf/internal/flags"
	"dario.lol/cf/internal/pagination"
	"dario.lol/cf/internal/ui"
	"dario.lol/cf/internal/ui/response"
	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/spf13/cobra"
)

var membersKey = executor.NewKey[[]accounts.Member]("members")

var membersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage account members",
}

var membersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List members of the current account",
	Run: executor.New().
		WithClient().
		WithAccountID().
		WithPagination().
		Step(executor.NewStep(membersKey, "Fetching members").Func(fetchMembers)).
		Display(printMembersList).
		Run(),
}

func init() {
	flags.RegisterAccountID(membersCmd)
	pagination.RegisterFlags(membersListCmd)
	membersCmd.AddCommand(membersListCmd)
	AccountCmd.AddCommand(membersCmd)
}

func fetchMembers(ctx *executor.Context, _ chan<- string) ([]accounts.Member, error) {
	membersList, err := ctx.Client.Accounts.Members.List(context.Background(), accounts.MemberListParams{
		AccountID: cf.F(ctx.AccountID),
	})
	if err != nil {
		return nil, err
	}
	return membersList.Result, nil
}

func printMembersList(ctx *executor.Context) {
	rb := response.New().Title("Account Members")

	if ctx.Error != nil {
		rb.Error("Error fetching members", ctx.Error).Display()
		return
	}

	members := executor.Get(ctx, membersKey)
	paginated, info := pagination.Paginate(members, ctx.Pagination)

	rb.Summary("Total:", info.Total)

	for _, member := range paginated {
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

	if len(paginated) > 0 {
		footer := fmt.Sprintf("Showing %d of %d member(s)", info.Showing, info.Total)
		footer += " " + ui.Muted(fmt.Sprintf("(took %v)", ctx.Duration))
		rb.FooterSuccess(footer)
	} else {
		rb.NoItemsMessage("No members found")
	}

	rb.Display()
}
