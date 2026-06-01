package cli

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/spf13/cobra"
)

type RepositoryUseCase interface {
	ListRepositories(ctx context.Context) ([]app.RepositoryListItem, error)
	GetRepository(ctx context.Context, id string) (app.RepositoryDetail, error)
	CloneRepository(ctx context.Context, id, destination string) error
}

func newRepoCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage Codeup repositories",
	}
	cmd.AddCommand(newRepoListCommand(opts))
	cmd.AddCommand(newRepoViewCommand(opts))
	cmd.AddCommand(newRepoCloneCommand(opts))
	return cmd
}

func newRepoListCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.repoUseCase()
			if err != nil {
				return err
			}
			repos, err := useCase.ListRepositories(cmd.Context())
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(repos)
			}
			rows := make([][]string, 0, len(repos))
			for _, repo := range repos {
				rows = append(rows, []string{repo.ID, repo.Name, repo.Path})
			}
			return renderer.WriteTable([]string{"ID", "NAME", "PATH"}, rows)
		},
	}
}

func newRepoViewCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "view <repo>",
		Short: "View a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.repoUseCase()
			if err != nil {
				return err
			}
			repo, err := useCase.GetRepository(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(repo)
			}
			return renderer.WriteTable(
				[]string{"ID", "NAME", "PATH", "CLONE_URL"},
				[][]string{{repo.ID, repo.Name, repo.Path, repo.CloneURL}},
			)
		},
	}
}

func newRepoCloneCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "clone <repo> [destination]",
		Short: "Clone a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.repoUseCase()
			if err != nil {
				return err
			}
			destination := ""
			if len(args) > 1 {
				destination = args[1]
			}
			return useCase.CloneRepository(cmd.Context(), args[0], destination)
		},
	}
}

func (o Options) repoUseCase() (RepositoryUseCase, error) {
	if o.RepoUseCase != nil {
		return o.RepoUseCase, nil
	}
	return nil, fmt.Errorf("repository service is not configured")
}
