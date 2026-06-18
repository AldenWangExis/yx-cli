package cli

import (
	"context"
	"strings"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/spf13/cobra"
)

type MemberUseCase interface {
	ListMembers(ctx context.Context, input app.MemberListInput) ([]app.Member, error)
	SearchMembers(ctx context.Context, input app.MemberSearchInput) ([]app.Member, error)
	GetMember(ctx context.Context, input app.MemberGetInput) (app.Member, error)
}

func newMemberCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "member",
		Short:   "Manage Yunxiao organization members",
		Long:    "Manage Yunxiao organization members.",
		Example: "  yx member list\n  yx member search --name 王子豪\n  yx member get --user-id <user-id>",
	}
	cmd.AddCommand(newMemberListCommand(opts))
	cmd.AddCommand(newMemberSearchCommand(opts))
	cmd.AddCommand(newMemberGetCommand(opts))
	return cmd
}

func newMemberListCommand(opts Options) *cobra.Command {
	var input app.MemberListInput
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List organization members",
		Example: "  yx member list\n  yx member list --status ENABLED\n  yx --json member list",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.memberUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			members, err := useCase.ListMembers(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderMembers(cmd, members)
		},
	}
	cmd.Flags().StringVar(&input.Status, "status", "", "member status")
	return cmd
}

func newMemberSearchCommand(opts Options) *cobra.Command {
	var input app.MemberSearchInput
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search organization members",
		Example: "  yx member search --name 王子豪\n  yx member search --email user@example.com",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.memberUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			members, err := useCase.SearchMembers(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderMembers(cmd, members)
		},
	}
	cmd.Flags().StringVar(&input.Name, "name", "", "member name")
	cmd.Flags().StringVar(&input.Email, "email", "", "member email")
	cmd.Flags().StringVar(&input.Status, "status", "", "member status")
	return cmd
}

func newMemberGetCommand(opts Options) *cobra.Command {
	var input app.MemberGetInput
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get an organization member by user ID",
		Example: "  yx member get --user-id <user-id>\n  yx --json member get --user-id <user-id>",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.memberUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			member, err := useCase.GetMember(cmd.Context(), input)
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(member)
			}
			return renderer.WriteTable([]string{"USER_ID", "NAME", "EMAIL", "STATUS", "ROLES"}, [][]string{memberRow(member)})
		},
	}
	cmd.Flags().StringVar(&input.UserID, "user-id", "", "member user id")
	return cmd
}

func renderMembers(cmd *cobra.Command, members []app.Member) error {
	renderer := output.NewRenderer(cmd.OutOrStdout())
	if ContextFromCommand(cmd).JSON {
		return renderer.WriteJSON(members)
	}
	rows := make([][]string, 0, len(members))
	for _, member := range members {
		rows = append(rows, memberRow(member))
	}
	return renderer.WriteTable([]string{"USER_ID", "NAME", "EMAIL", "STATUS", "ROLES"}, rows)
}

func memberRow(member app.Member) []string {
	return []string{member.UserID, member.Name, member.Email, member.Status, strings.Join(member.RoleIDs, ",")}
}

func (o Options) memberUseCase(ctx Context) (MemberUseCase, error) {
	if o.MemberUseCase != nil {
		return o.MemberUseCase, nil
	}
	services, err := o.resolveRuntimeServices(ctx)
	if err != nil {
		return nil, err
	}
	return services.memberUseCase(), nil
}
