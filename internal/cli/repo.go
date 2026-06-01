package cli

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/AldenWangExis/yx-cli/internal/gitx"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/AldenWangExis/yx-cli/internal/safety"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/codeup"
	"github.com/spf13/cobra"
)

type RepositoryUseCase interface {
	ListRepositories(ctx context.Context) ([]app.RepositoryListItem, error)
	GetRepository(ctx context.Context, id string) (app.RepositoryDetail, error)
	CreateRepository(ctx context.Context, input app.CreateRepositoryInput) (app.RepositoryMutationResult, error)
	CloneRepository(ctx context.Context, id, destination string) error
	ListBranches(ctx context.Context, repo string) ([]app.BranchListItem, error)
	SyncBranch(ctx context.Context, input app.BranchSyncInput) (app.BranchMutationResult, error)
	ListCommits(ctx context.Context, input app.CommitListInput) ([]app.CommitListItem, error)
	GetFile(ctx context.Context, input app.FileGetInput) (app.RepositoryFile, error)
}

func newRepoCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage Codeup repositories",
	}
	cmd.AddCommand(newRepoListCommand(opts))
	cmd.AddCommand(newRepoViewCommand(opts))
	cmd.AddCommand(newRepoCreateCommand(opts))
	cmd.AddCommand(newRepoCloneCommand(opts))
	cmd.AddCommand(newRepoBranchCommand(opts))
	cmd.AddCommand(newRepoCommitCommand(opts))
	cmd.AddCommand(newRepoFileCommand(opts))
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

func newRepoCreateCommand(opts Options) *cobra.Command {
	input := app.CreateRepositoryInput{Visibility: "private", ReadmeType: "EMPTY"}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Codeup repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.repoUseCase()
			if err != nil {
				return err
			}
			result, err := useCase.CreateRepository(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderRepositoryMutation(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.Name, "name", "", "repository name")
	cmd.Flags().StringVar(&input.Path, "path", "", "repository path")
	cmd.Flags().StringVar(&input.Description, "description", "", "repository description")
	cmd.Flags().StringVar(&input.Visibility, "visibility", "private", "repository visibility")
	cmd.Flags().StringVar(&input.ReadmeType, "readme-type", "EMPTY", "repository readme initialization type")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
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

func newRepoBranchCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{Use: "branch", Short: "Manage repository branches"}
	cmd.AddCommand(newRepoBranchListCommand(opts))
	cmd.AddCommand(newRepoBranchSyncCommand(opts))
	return cmd
}

func newRepoBranchListCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list <repo>",
		Short: "List repository branches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.repoUseCase()
			if err != nil {
				return err
			}
			branches, err := useCase.ListBranches(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(branches)
			}
			rows := make([][]string, 0, len(branches))
			for _, branch := range branches {
				rows = append(rows, []string{branch.Name, fmt.Sprintf("%t", branch.Default), branch.CommitID})
			}
			return renderer.WriteTable([]string{"NAME", "DEFAULT", "COMMIT"}, rows)
		},
	}
}

func newRepoBranchSyncCommand(opts Options) *cobra.Command {
	var input app.BranchSyncInput
	cmd := &cobra.Command{
		Use:   "sync <repo>",
		Short: "Create a remote branch from a source ref",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input.Repo = args[0]
			useCase, err := opts.repoUseCase()
			if err != nil {
				return err
			}
			result, err := useCase.SyncBranch(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderBranchMutation(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.Source, "source", "", "source branch or ref")
	cmd.Flags().StringVar(&input.Target, "target", "", "target branch to create")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newRepoCommitCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{Use: "commit", Short: "Inspect repository commits"}
	cmd.AddCommand(newRepoCommitListCommand(opts))
	return cmd
}

func newRepoCommitListCommand(opts Options) *cobra.Command {
	var input app.CommitListInput
	cmd := &cobra.Command{
		Use:   "list <repo>",
		Short: "List repository commits",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input.Repo = args[0]
			useCase, err := opts.repoUseCase()
			if err != nil {
				return err
			}
			commits, err := useCase.ListCommits(cmd.Context(), input)
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(commits)
			}
			rows := make([][]string, 0, len(commits))
			for _, commit := range commits {
				rows = append(rows, []string{commit.ShortID, commit.Title, commit.AuthorName, commit.CommittedDate})
			}
			return renderer.WriteTable([]string{"COMMIT", "TITLE", "AUTHOR", "DATE"}, rows)
		},
	}
	cmd.Flags().StringVar(&input.Ref, "ref", "", "branch or ref")
	return cmd
}

func newRepoFileCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{Use: "file", Short: "Inspect repository files"}
	cmd.AddCommand(newRepoFileViewCommand(opts))
	return cmd
}

func newRepoFileViewCommand(opts Options) *cobra.Command {
	var input app.FileGetInput
	cmd := &cobra.Command{
		Use:   "view <repo> <path>",
		Short: "View a repository file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			input.Repo = args[0]
			input.Path = args[1]
			useCase, err := opts.repoUseCase()
			if err != nil {
				return err
			}
			file, err := useCase.GetFile(cmd.Context(), input)
			if err != nil {
				return err
			}
			if ContextFromCommand(cmd).JSON {
				return output.NewRenderer(cmd.OutOrStdout()).WriteJSON(file)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), file.Content)
			return err
		},
	}
	cmd.Flags().StringVar(&input.Ref, "ref", "", "branch or ref")
	return cmd
}

func renderRepositoryMutation(cmd *cobra.Command, result app.RepositoryMutationResult) error {
	renderer := output.NewRenderer(cmd.OutOrStdout())
	if ContextFromCommand(cmd).JSON {
		return renderer.WriteJSON(result)
	}
	if result.DryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %s\n", result.Summary)
		return nil
	}
	repo := result.Repository
	return renderer.WriteTable([]string{"ID", "NAME", "PATH", "WEB_URL"}, [][]string{{repo.ID, repo.Name, repo.Path, repo.WebURL}})
}

func renderBranchMutation(cmd *cobra.Command, result app.BranchMutationResult) error {
	renderer := output.NewRenderer(cmd.OutOrStdout())
	if ContextFromCommand(cmd).JSON {
		return renderer.WriteJSON(result)
	}
	if result.DryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %s\n", result.Summary)
		return nil
	}
	branch := result.Branch
	return renderer.WriteTable([]string{"NAME", "COMMIT"}, [][]string{{branch.Name, branch.CommitID}})
}

func (o Options) repoUseCase() (RepositoryUseCase, error) {
	if o.RepoUseCase != nil {
		return o.RepoUseCase, nil
	}
	cfg, err := config.NewStore(o.ConfigPath).Load()
	if err != nil {
		return nil, err
	}
	profileName := o.DefaultProfile
	if profileName == "" {
		profileName = cfg.Current
	}
	if profileName == "" {
		profileName = "default"
	}
	profile, ok := cfg.Profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile %q does not exist", profileName)
	}
	token, ok, err := auth.NewFileTokenStore(defaultTokenPath(o.ConfigPath)).Load(profileName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("profile %q is not logged in", profileName)
	}
	repositories := codeup.NewRepositoryAdapter(yunxiao.ClientConfig{
		BaseURL:        profile.Domain,
		Token:          token,
		OrganizationID: profile.Organization,
		Region:         profile.Region,
	})
	return app.NewRepoUseCase(repositories, gitx.NewRunner(), safety.Environment{
		ConfirmWrites: profile.Safety.ConfirmWrites,
	}), nil
}
