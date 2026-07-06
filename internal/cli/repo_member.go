package cli

import (
	"fmt"
	"strconv"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/spf13/cobra"
)

func newRepoMemberCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "member",
		Short:   "Manage repository members",
		Long:    "Manage Codeup repository members and access levels.",
		Example: "  yx repo member list <repo>\n  yx repo member add <repo> --user-id <user-id> --access-level developer --dry-run\n  yx repo member update <repo> --user-id <user-id> --access-level maintainer --dry-run\n  yx repo member remove <repo> --user-id <user-id> --dry-run",
	}
	cmd.AddCommand(newRepoMemberListCommand(opts))
	cmd.AddCommand(newRepoMemberAddCommand(opts))
	cmd.AddCommand(newRepoMemberUpdateCommand(opts))
	cmd.AddCommand(newRepoMemberRemoveCommand(opts))
	return cmd
}

func newRepoMemberListCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [repo]",
		Short:   "List repository members",
		Example: "  yx repo member list\n  yx repo member list 6925595\n  yx --json repo member list 6925595",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := repoArgumentOrCurrent(cmd, opts, args)
			if err != nil {
				return err
			}
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			members, err := useCase.ListRepositoryMembers(cmd.Context(), repo)
			if err != nil {
				return err
			}
			return renderRepositoryMembers(cmd, members)
		},
	}
	return cmd
}

func newRepoMemberAddCommand(opts Options) *cobra.Command {
	var input app.AddRepositoryMemberInput
	cmd := &cobra.Command{
		Use:     "add [repo]",
		Short:   "Add a repository member",
		Example: "  yx repo member add 6925595 --user-id <user-id> --access-level developer --dry-run",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := repoArgumentOrCurrent(cmd, opts, args)
			if err != nil {
				return err
			}
			input.Repo = repo
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			result, err := useCase.AddRepositoryMember(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderRepositoryMemberMutation(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.UserID, "user-id", "", "member user id")
	cmd.Flags().StringVar(&input.AccessLevel, "access-level", "", "access level: viewer, developer, maintainer, 20, 30, or 40")
	cmd.Flags().StringVar(&input.ExpiresAt, "expires-at", "", "optional access expiration date")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newRepoMemberUpdateCommand(opts Options) *cobra.Command {
	var input app.UpdateRepositoryMemberInput
	cmd := &cobra.Command{
		Use:     "update [repo]",
		Short:   "Update repository member access",
		Example: "  yx repo member update 6925595 --user-id <user-id> --access-level maintainer --dry-run",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := repoArgumentOrCurrent(cmd, opts, args)
			if err != nil {
				return err
			}
			input.Repo = repo
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			result, err := useCase.UpdateRepositoryMember(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderRepositoryMemberMutation(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.UserID, "user-id", "", "member user id")
	cmd.Flags().StringVar(&input.AccessLevel, "access-level", "", "access level: viewer, developer, maintainer, 20, 30, or 40")
	cmd.Flags().StringVar(&input.ExpiresAt, "expires-at", "", "optional access expiration date")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newRepoMemberRemoveCommand(opts Options) *cobra.Command {
	var input app.RemoveRepositoryMemberInput
	cmd := &cobra.Command{
		Use:     "remove [repo]",
		Short:   "Remove a repository member",
		Aliases: []string{"delete"},
		Example: "  yx repo member remove 6925595 --user-id <user-id> --dry-run",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := repoArgumentOrCurrent(cmd, opts, args)
			if err != nil {
				return err
			}
			input.Repo = repo
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			result, err := useCase.RemoveRepositoryMember(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderRepositoryMemberMutation(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.UserID, "user-id", "", "member user id")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func repoArgumentOrCurrent(cmd *cobra.Command, opts Options, args []string) (string, error) {
	repo := ""
	if len(args) > 0 {
		repo = args[0]
	}
	return resolveRepositoryID(cmd, opts, repo)
}

func renderRepositoryMembers(cmd *cobra.Command, members []app.RepositoryMember) error {
	renderer := output.NewRenderer(cmd.OutOrStdout())
	if ContextFromCommand(cmd).JSON {
		return renderer.WriteJSON(members)
	}
	rows := make([][]string, 0, len(members))
	for _, member := range members {
		rows = append(rows, repositoryMemberRow(member))
	}
	return renderer.WriteTable([]string{"USER_ID", "NAME", "EMAIL", "ACCESS", "EXPIRES_AT", "INHERITED", "SOURCE"}, rows)
}

func renderRepositoryMemberMutation(cmd *cobra.Command, result app.RepositoryMemberMutationResult) error {
	renderer := output.NewRenderer(cmd.OutOrStdout())
	if ContextFromCommand(cmd).JSON {
		return renderer.WriteJSON(result)
	}
	if result.DryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %s\n", result.Summary)
		return nil
	}
	return renderer.WriteTable([]string{"USER_ID", "NAME", "EMAIL", "ACCESS", "EXPIRES_AT", "INHERITED", "SOURCE"}, [][]string{repositoryMemberRow(result.Member)})
}

func repositoryMemberRow(member app.RepositoryMember) []string {
	access := member.Access
	if access == "" {
		access = app.RepositoryAccessLevelName(member.AccessLevel)
	}
	inherited := strconv.FormatBool(member.Inherited)
	return []string{member.UserID, member.Name, member.Email, access, member.ExpiresAt, inherited, member.Source}
}
