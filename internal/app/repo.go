package app

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

type RepositoryListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

type RepositoryDetail struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"path,omitempty"`
	CloneURL      string `json:"cloneUrl"`
	WebURL        string `json:"webUrl,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
}

type CreateRepositoryInput struct {
	Name        string
	Path        string
	Description string
	Visibility  string
	ReadmeType  string
	DryRun      bool
	Yes         bool
}

type RepositoryMutationResult struct {
	DryRun     bool             `json:"dryRun"`
	Summary    string           `json:"summary,omitempty"`
	Repository RepositoryDetail `json:"repository,omitempty"`
}

type BranchListItem struct {
	Name      string `json:"name"`
	Default   bool   `json:"default"`
	Protected bool   `json:"protected"`
	CommitID  string `json:"commitId,omitempty"`
	WebURL    string `json:"webUrl,omitempty"`
}

type BranchSyncInput struct {
	Repo   string
	Source string
	Target string
	DryRun bool
	Yes    bool
}

type BranchMutationResult struct {
	DryRun  bool           `json:"dryRun"`
	Summary string         `json:"summary,omitempty"`
	Branch  BranchListItem `json:"branch,omitempty"`
}

type CommitListInput struct {
	Repo string
	Ref  string
}

type CommitListItem struct {
	ID            string `json:"id"`
	ShortID       string `json:"shortId,omitempty"`
	Title         string `json:"title"`
	Message       string `json:"message,omitempty"`
	AuthorName    string `json:"authorName,omitempty"`
	CommittedDate string `json:"committedDate,omitempty"`
	WebURL        string `json:"webUrl,omitempty"`
}

type FileGetInput struct {
	Repo string
	Path string
	Ref  string
}

type RepositoryFile struct {
	Path     string `json:"path"`
	Ref      string `json:"ref"`
	Encoding string `json:"encoding,omitempty"`
	Content  string `json:"content"`
}

type RepositoryService interface {
	ListRepositories(ctx context.Context) ([]RepositoryListItem, error)
	GetRepository(ctx context.Context, id string) (RepositoryDetail, error)
	CreateRepository(ctx context.Context, input CreateRepositoryInput) (RepositoryDetail, error)
	ListBranches(ctx context.Context, repo string) ([]BranchListItem, error)
	SyncBranch(ctx context.Context, input BranchSyncInput) (BranchListItem, error)
	ListCommits(ctx context.Context, input CommitListInput) ([]CommitListItem, error)
	GetFile(ctx context.Context, input FileGetInput) (RepositoryFile, error)
}

type GitRunner interface {
	Clone(ctx context.Context, cloneURL, destination string) error
}

type RepoUseCase struct {
	repositories RepositoryService
	git          GitRunner
	safety       safety.Environment
}

func NewRepoUseCase(repositories RepositoryService, git GitRunner, safetyEnv safety.Environment) *RepoUseCase {
	return &RepoUseCase{repositories: repositories, git: git, safety: safetyEnv}
}

func (u *RepoUseCase) ListRepositories(ctx context.Context) ([]RepositoryListItem, error) {
	return u.repositories.ListRepositories(ctx)
}

func (u *RepoUseCase) GetRepository(ctx context.Context, id string) (RepositoryDetail, error) {
	return u.repositories.GetRepository(ctx, id)
}

func (u *RepoUseCase) CreateRepository(ctx context.Context, input CreateRepositoryInput) (RepositoryMutationResult, error) {
	if input.Name == "" {
		return RepositoryMutationResult{}, fmt.Errorf("name is required")
	}
	if input.Path == "" {
		input.Path = input.Name
	}
	if input.Visibility == "" {
		input.Visibility = "private"
	}
	if input.ReadmeType == "" {
		input.ReadmeType = "EMPTY"
	}
	summary := fmt.Sprintf("create repository %q at %s", input.Name, input.Path)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return RepositoryMutationResult{}, err
	}
	if decision.DryRun {
		return RepositoryMutationResult{DryRun: true, Summary: summary}, nil
	}
	repo, err := u.repositories.CreateRepository(ctx, input)
	if err != nil {
		return RepositoryMutationResult{}, err
	}
	if repo.Name == "" {
		repo.Name = input.Name
	}
	if repo.Path == "" {
		repo.Path = input.Path
	}
	return RepositoryMutationResult{Repository: repo}, nil
}

func (u *RepoUseCase) CloneRepository(ctx context.Context, id, destination string) error {
	repo, err := u.repositories.GetRepository(ctx, id)
	if err != nil {
		return err
	}
	if repo.CloneURL == "" {
		return fmt.Errorf("repository %q does not include a clone URL", id)
	}
	return u.git.Clone(ctx, repo.CloneURL, destination)
}

func (u *RepoUseCase) ListBranches(ctx context.Context, repo string) ([]BranchListItem, error) {
	if repo == "" {
		return nil, fmt.Errorf("repo is required")
	}
	return u.repositories.ListBranches(ctx, repo)
}

func (u *RepoUseCase) SyncBranch(ctx context.Context, input BranchSyncInput) (BranchMutationResult, error) {
	if input.Repo == "" || input.Source == "" || input.Target == "" {
		return BranchMutationResult{}, fmt.Errorf("repo, source, and target are required")
	}
	summary := fmt.Sprintf("sync branch %s from %s in %s", input.Target, input.Source, input.Repo)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return BranchMutationResult{}, err
	}
	if decision.DryRun {
		return BranchMutationResult{DryRun: true, Summary: summary}, nil
	}
	branch, err := u.repositories.SyncBranch(ctx, input)
	if err != nil {
		return BranchMutationResult{}, err
	}
	return BranchMutationResult{Branch: branch}, nil
}

func (u *RepoUseCase) ListCommits(ctx context.Context, input CommitListInput) ([]CommitListItem, error) {
	if input.Repo == "" {
		return nil, fmt.Errorf("repo is required")
	}
	return u.repositories.ListCommits(ctx, input)
}

func (u *RepoUseCase) GetFile(ctx context.Context, input FileGetInput) (RepositoryFile, error) {
	if input.Repo == "" || input.Path == "" {
		return RepositoryFile{}, fmt.Errorf("repo and path are required")
	}
	if input.Ref == "" {
		input.Ref = "master"
	}
	return u.repositories.GetFile(ctx, input)
}
