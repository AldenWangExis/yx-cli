package app

import (
	"context"
	"fmt"
	"strings"

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

type DeleteRepositoryInput struct {
	Repo   string
	DryRun bool
	Yes    bool
}

type RepositoryMutationResult struct {
	DryRun     bool             `json:"dryRun"`
	Summary    string           `json:"summary,omitempty"`
	Repository RepositoryDetail `json:"repository,omitempty"`
}

type RepositoryMember struct {
	UserID      string `json:"userId"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	AccessLevel int    `json:"accessLevel,omitempty"`
	Access      string `json:"access,omitempty"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
	Inherited   bool   `json:"inherited,omitempty"`
	Source      string `json:"source,omitempty"`
}

type AddRepositoryMemberInput struct {
	Repo        string
	UserID      string
	AccessLevel string
	ExpiresAt   string
	DryRun      bool
	Yes         bool
}

type UpdateRepositoryMemberInput struct {
	Repo        string
	UserID      string
	AccessLevel string
	ExpiresAt   string
	DryRun      bool
	Yes         bool
}

type RemoveRepositoryMemberInput struct {
	Repo   string
	UserID string
	DryRun bool
	Yes    bool
}

type RepositoryMemberMutationResult struct {
	DryRun  bool             `json:"dryRun"`
	Summary string           `json:"summary,omitempty"`
	Member  RepositoryMember `json:"member,omitempty"`
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
	DeleteRepository(ctx context.Context, repo string) (RepositoryDetail, error)
	ListRepositoryMembers(ctx context.Context, repo string) ([]RepositoryMember, error)
	AddRepositoryMember(ctx context.Context, input AddRepositoryMemberInput) (RepositoryMember, error)
	UpdateRepositoryMember(ctx context.Context, input UpdateRepositoryMemberInput) (RepositoryMember, error)
	RemoveRepositoryMember(ctx context.Context, input RemoveRepositoryMemberInput) (RepositoryMember, error)
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

func (u *RepoUseCase) DeleteRepository(ctx context.Context, input DeleteRepositoryInput) (RepositoryMutationResult, error) {
	if input.Repo == "" {
		return RepositoryMutationResult{}, fmt.Errorf("repo is required")
	}
	summary := fmt.Sprintf("delete repository %s", input.Repo)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return RepositoryMutationResult{}, err
	}
	if decision.DryRun {
		return RepositoryMutationResult{DryRun: true, Summary: summary}, nil
	}
	repo, err := u.repositories.DeleteRepository(ctx, input.Repo)
	if err != nil {
		return RepositoryMutationResult{}, err
	}
	if repo.ID == "" {
		repo.ID = input.Repo
	}
	return RepositoryMutationResult{Repository: repo}, nil
}

func (u *RepoUseCase) ListRepositoryMembers(ctx context.Context, repo string) ([]RepositoryMember, error) {
	if repo == "" {
		return nil, fmt.Errorf("repo is required")
	}
	return u.repositories.ListRepositoryMembers(ctx, repo)
}

func (u *RepoUseCase) AddRepositoryMember(ctx context.Context, input AddRepositoryMemberInput) (RepositoryMemberMutationResult, error) {
	if input.Repo == "" || input.UserID == "" || input.AccessLevel == "" {
		return RepositoryMemberMutationResult{}, fmt.Errorf("repo, user id, and access level are required")
	}
	accessLevel, err := NormalizeRepositoryAccessLevel(input.AccessLevel)
	if err != nil {
		return RepositoryMemberMutationResult{}, err
	}
	input.AccessLevel = accessLevel
	summary := fmt.Sprintf("add repository member %s to %s as %s", input.UserID, input.Repo, input.AccessLevel)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return RepositoryMemberMutationResult{}, err
	}
	if decision.DryRun {
		return RepositoryMemberMutationResult{DryRun: true, Summary: summary}, nil
	}
	member, err := u.repositories.AddRepositoryMember(ctx, input)
	if err != nil {
		return RepositoryMemberMutationResult{}, err
	}
	if member.UserID == "" {
		member.UserID = input.UserID
	}
	return RepositoryMemberMutationResult{Member: member}, nil
}

func (u *RepoUseCase) UpdateRepositoryMember(ctx context.Context, input UpdateRepositoryMemberInput) (RepositoryMemberMutationResult, error) {
	if input.Repo == "" || input.UserID == "" || input.AccessLevel == "" {
		return RepositoryMemberMutationResult{}, fmt.Errorf("repo, user id, and access level are required")
	}
	accessLevel, err := NormalizeRepositoryAccessLevel(input.AccessLevel)
	if err != nil {
		return RepositoryMemberMutationResult{}, err
	}
	input.AccessLevel = accessLevel
	summary := fmt.Sprintf("update repository member %s in %s to %s", input.UserID, input.Repo, input.AccessLevel)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return RepositoryMemberMutationResult{}, err
	}
	if decision.DryRun {
		return RepositoryMemberMutationResult{DryRun: true, Summary: summary}, nil
	}
	member, err := u.repositories.UpdateRepositoryMember(ctx, input)
	if err != nil {
		return RepositoryMemberMutationResult{}, err
	}
	if member.UserID == "" {
		member.UserID = input.UserID
	}
	return RepositoryMemberMutationResult{Member: member}, nil
}

func (u *RepoUseCase) RemoveRepositoryMember(ctx context.Context, input RemoveRepositoryMemberInput) (RepositoryMemberMutationResult, error) {
	if input.Repo == "" || input.UserID == "" {
		return RepositoryMemberMutationResult{}, fmt.Errorf("repo and user id are required")
	}
	summary := fmt.Sprintf("remove repository member %s from %s", input.UserID, input.Repo)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return RepositoryMemberMutationResult{}, err
	}
	if decision.DryRun {
		return RepositoryMemberMutationResult{DryRun: true, Summary: summary}, nil
	}
	member, err := u.repositories.RemoveRepositoryMember(ctx, input)
	if err != nil {
		return RepositoryMemberMutationResult{}, err
	}
	if member.UserID == "" {
		member.UserID = input.UserID
	}
	return RepositoryMemberMutationResult{Member: member}, nil
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

func NormalizeRepositoryAccessLevel(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "20", "viewer", "guest", "reporter":
		return "20", nil
	case "30", "developer":
		return "30", nil
	case "40", "maintainer", "admin":
		return "40", nil
	default:
		return "", fmt.Errorf("access level must be viewer, developer, maintainer, 20, 30, or 40")
	}
}

func RepositoryAccessLevelName(level int) string {
	switch level {
	case 20:
		return "viewer"
	case 30:
		return "developer"
	case 40:
		return "maintainer"
	default:
		if level == 0 {
			return ""
		}
		return fmt.Sprintf("%d", level)
	}
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
