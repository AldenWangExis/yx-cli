package app

import (
	"context"
	"fmt"
)

type RepositoryListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

type RepositoryDetail struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
	CloneURL string `json:"cloneUrl"`
}

type RepositoryService interface {
	ListRepositories(ctx context.Context) ([]RepositoryListItem, error)
	GetRepository(ctx context.Context, id string) (RepositoryDetail, error)
}

type GitRunner interface {
	Clone(ctx context.Context, cloneURL, destination string) error
}

type RepoUseCase struct {
	repositories RepositoryService
	git          GitRunner
}

func NewRepoUseCase(repositories RepositoryService, git GitRunner) *RepoUseCase {
	return &RepoUseCase{repositories: repositories, git: git}
}

func (u *RepoUseCase) ListRepositories(ctx context.Context) ([]RepositoryListItem, error) {
	return u.repositories.ListRepositories(ctx)
}

func (u *RepoUseCase) GetRepository(ctx context.Context, id string) (RepositoryDetail, error) {
	return u.repositories.GetRepository(ctx, id)
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
