package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/spf13/cobra"
)

type RepositoryUseCase interface {
	ListRepositories(ctx context.Context) ([]app.RepositoryListItem, error)
	GetRepository(ctx context.Context, id string) (app.RepositoryDetail, error)
	CreateRepository(ctx context.Context, input app.CreateRepositoryInput) (app.RepositoryMutationResult, error)
	CloneRepository(ctx context.Context, id, destination string) error
	DeleteRepository(ctx context.Context, input app.DeleteRepositoryInput) (app.RepositoryMutationResult, error)
	ListBranches(ctx context.Context, repo string) ([]app.BranchListItem, error)
	SyncBranch(ctx context.Context, input app.BranchSyncInput) (app.BranchMutationResult, error)
	ListCommits(ctx context.Context, input app.CommitListInput) ([]app.CommitListItem, error)
	GetFile(ctx context.Context, input app.FileGetInput) (app.RepositoryFile, error)
}

type RepositoryCurrentResolver interface {
	CurrentRepository(ctx context.Context, input app.CurrentRepositoryInput) (app.CurrentRepository, error)
}

func newRepoCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repo",
		Short:   "Manage Codeup repositories",
		Long:    "Manage Codeup repositories, branches, commits, files, and clones.",
		Example: "  yx repo list\n  yx repo current\n  yx repo view\n  yx repo view <repo>\n  yx repo create --name demo --path demo --visibility private --yes\n  yx repo delete <repo> --dry-run\n  yx repo branch list\n  yx repo commit list --ref master\n  yx repo file view test.py --ref master",
	}
	cmd.AddCommand(newRepoListCommand(opts))
	cmd.AddCommand(newRepoCurrentCommand(opts))
	cmd.AddCommand(newRepoViewCommand(opts))
	cmd.AddCommand(newRepoCreateCommand(opts))
	cmd.AddCommand(newRepoCloneCommand(opts))
	cmd.AddCommand(newRepoDeleteCommand(opts))
	cmd.AddCommand(newRepoBranchCommand(opts))
	cmd.AddCommand(newRepoCommitCommand(opts))
	cmd.AddCommand(newRepoFileCommand(opts))
	return cmd
}

func newRepoDeleteCommand(opts Options) *cobra.Command {
	var input app.DeleteRepositoryInput
	cmd := &cobra.Command{
		Use:     "delete [repo]",
		Short:   "Delete a Codeup repository",
		Example: "  yx repo delete --dry-run\n  yx repo delete 6925595 --yes",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				input.Repo = args[0]
			}
			var err error
			input.Repo, err = resolveRepositoryID(cmd, opts, input.Repo)
			if err != nil {
				return err
			}
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			result, err := useCase.DeleteRepository(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderRepositoryMutation(cmd, result)
		},
	}
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newRepoCurrentCommand(opts Options) *cobra.Command {
	var remote string
	var refresh bool
	cmd := &cobra.Command{
		Use:     "current",
		Short:   "Resolve the current Codeup repository from git remotes",
		Long:    "Resolve the current Codeup repository by reading git remotes, matching the exact Codeup path through OpenAPI, and caching the repository identity.",
		Example: "  yx repo current\n  yx --json repo current\n  yx repo current --remote codeup\n  yx repo current --refresh",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolver, runtimeContext, err := opts.repoCurrentResolver(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			workDir, err := opts.currentWorkDir()
			if err != nil {
				return err
			}
			current, err := resolver.CurrentRepository(cmd.Context(), app.CurrentRepositoryInput{
				ProfileName:  runtimeContext.ProfileName,
				Domain:       runtimeContext.Domain,
				Organization: runtimeContext.Organization,
				Region:       runtimeContext.Region,
				WorkDir:      workDir,
				Remote:       remote,
				Refresh:      refresh,
			})
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(current)
			}
			return renderer.WriteTable(
				[]string{"ID", "NAME", "PATH", "REMOTE", "SOURCE"},
				[][]string{{string(current.ID), current.Name, string(current.Path), string(current.Remote), current.Source}},
			)
		},
	}
	cmd.Flags().StringVar(&remote, "remote", "", "git remote name to resolve")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "bypass cached repository identity")
	return cmd
}

func newRepoListCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
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
		Use:   "view [repo]",
		Short: "View a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoID := ""
			if len(args) > 0 {
				repoID = args[0]
			}
			repoID, err := resolveRepositoryID(cmd, opts, repoID)
			if err != nil {
				return err
			}
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			repo, err := useCase.GetRepository(cmd.Context(), repoID)
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
		Use:     "create",
		Short:   "Create a Codeup repository",
		Long:    "Create a Codeup repository through Yunxiao OpenAPI.",
		Example: "  yx repo create --name demo --path demo --visibility private --yes\n  yx repo create --name demo --dry-run",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
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
		Use:     "clone <repo> [destination]",
		Short:   "Clone a repository",
		Example: "  yx repo clone 6925595\n  yx repo clone 6925595 ./repo-dir",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
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
	cmd := &cobra.Command{
		Use:     "branch",
		Short:   "Manage repository branches",
		Example: "  yx repo branch list\n  yx repo branch list <repo>\n  yx repo branch sync --source master --target feat/a --dry-run\n  yx repo branch sync <repo> --source master --target feat/a --dry-run",
	}
	cmd.AddCommand(newRepoBranchListCommand(opts))
	cmd.AddCommand(newRepoBranchSyncCommand(opts))
	return cmd
}

func newRepoBranchListCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:     "list [repo]",
		Short:   "List repository branches",
		Example: "  yx repo branch list\n  yx repo branch list 6925595\n  yx --json repo branch list",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoID := ""
			if len(args) > 0 {
				repoID = args[0]
			}
			repoID, err := resolveRepositoryID(cmd, opts, repoID)
			if err != nil {
				return err
			}
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			branches, err := useCase.ListBranches(cmd.Context(), repoID)
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
		Use:     "sync [repo]",
		Short:   "Create a remote branch from a source ref",
		Example: "  yx repo branch sync --source master --target feat/a --dry-run\n  yx repo branch sync 6925595 --source master --target feat/a --yes",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				input.Repo = args[0]
			}
			var err error
			input.Repo, err = resolveRepositoryID(cmd, opts, input.Repo)
			if err != nil {
				return err
			}
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
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
	cmd := &cobra.Command{
		Use:     "commit",
		Short:   "Inspect repository commits",
		Example: "  yx repo commit list --ref master\n  yx repo commit list <repo> --ref master",
	}
	cmd.AddCommand(newRepoCommitListCommand(opts))
	return cmd
}

func newRepoCommitListCommand(opts Options) *cobra.Command {
	var input app.CommitListInput
	cmd := &cobra.Command{
		Use:     "list [repo]",
		Short:   "List repository commits",
		Example: "  yx repo commit list --ref master\n  yx repo commit list 6925595 --ref master\n  yx --json repo commit list",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				input.Repo = args[0]
			}
			var err error
			input.Repo, err = resolveRepositoryID(cmd, opts, input.Repo)
			if err != nil {
				return err
			}
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
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
	cmd := &cobra.Command{
		Use:     "file",
		Short:   "Inspect repository files",
		Example: "  yx repo file view test.py --ref master\n  yx repo file view <repo> test.py --ref master",
	}
	cmd.AddCommand(newRepoFileViewCommand(opts))
	return cmd
}

func newRepoFileViewCommand(opts Options) *cobra.Command {
	var input app.FileGetInput
	cmd := &cobra.Command{
		Use:     "view [repo] <path>",
		Short:   "View a repository file",
		Example: "  yx repo file view test.py --ref master\n  yx repo file view 6925595 test.py --ref master\n  yx --json repo file view test.py --ref master",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				input.Path = args[0]
			} else {
				input.Repo = args[0]
				input.Path = args[1]
			}
			var err error
			input.Repo, err = resolveRepositoryID(cmd, opts, input.Repo)
			if err != nil {
				return err
			}
			useCase, err := opts.repoUseCase(ContextFromCommand(cmd))
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

func (o Options) repoUseCase(ctx Context) (RepositoryUseCase, error) {
	if o.RepoUseCase != nil {
		return o.RepoUseCase, nil
	}
	services, err := o.resolveRuntimeServices(ctx)
	if err != nil {
		return nil, err
	}
	return services.repoUseCase(), nil
}

type repoCurrentRuntimeContext struct {
	ProfileName  string
	Domain       string
	Organization string
	Region       string
}

func (o Options) repoCurrentResolver(ctx Context) (RepositoryCurrentResolver, repoCurrentRuntimeContext, error) {
	if o.RepoCurrentResolver != nil {
		profileName := ctx.Profile
		if profileName == "" {
			profileName = o.DefaultProfile
		}
		if profileName == "" {
			profileName = "default"
		}
		return o.RepoCurrentResolver, repoCurrentRuntimeContext{
			ProfileName:  profileName,
			Domain:       ctx.Domain,
			Organization: ctx.Organization,
		}, nil
	}
	services, err := o.resolveRuntimeServices(ctx)
	if err != nil {
		return nil, repoCurrentRuntimeContext{}, err
	}
	return services.repoCurrentResolver(), services.repoCurrentContext(), nil
}

type configRepositoryIdentityCache struct {
	store *config.Store
}

func (c configRepositoryIdentityCache) LookupRepositoryIdentity(key app.RepositoryIdentityCacheKey) (app.CurrentRepository, bool, error) {
	cfg, err := c.store.Load()
	if err != nil {
		return app.CurrentRepository{}, false, err
	}
	profile, ok := cfg.Profiles[key.ProfileName]
	if !ok {
		return app.CurrentRepository{}, false, nil
	}
	identity, ok := profile.RepoIdentityMap[key.StorageKey()]
	if !ok {
		return app.CurrentRepository{}, false, nil
	}
	if identity.Path != "" && identity.Path != key.Path {
		return app.CurrentRepository{}, false, nil
	}
	if identity.Domain != "" && identity.Domain != key.Domain {
		return app.CurrentRepository{}, false, nil
	}
	if identity.Organization != "" && identity.Organization != key.Organization {
		return app.CurrentRepository{}, false, nil
	}
	if identity.Region != "" && identity.Region != key.Region {
		return app.CurrentRepository{}, false, nil
	}
	return app.CurrentRepository{
		ID:           app.RepositoryID(identity.ID),
		Name:         identity.Name,
		Path:         app.RepositoryPath(identity.Path),
		Remote:       app.GitRemoteName(identity.Remote),
		Domain:       identity.Domain,
		Organization: identity.Organization,
		Region:       identity.Region,
	}, true, nil
}

func (c configRepositoryIdentityCache) StoreRepositoryIdentity(key app.RepositoryIdentityCacheKey, repo app.CurrentRepository) error {
	cfg, err := c.store.Load()
	if err != nil {
		return err
	}
	profile := cfg.Profiles[key.ProfileName]
	if profile.RepoIdentityMap == nil {
		profile.RepoIdentityMap = map[string]config.RepoIdentity{}
	}
	profile.RepoIdentityMap[key.StorageKey()] = config.RepoIdentity{
		ID:           string(repo.ID),
		Name:         repo.Name,
		Path:         string(repo.Path),
		Remote:       string(repo.Remote),
		Domain:       key.Domain,
		Organization: key.Organization,
		Region:       key.Region,
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	cfg.Profiles[key.ProfileName] = profile
	return c.store.Save(cfg)
}
