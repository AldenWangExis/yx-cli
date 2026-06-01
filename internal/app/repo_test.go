package app

import (
	"context"
	"errors"
	"testing"
)

func TestRepoUseCaseListsRepositories(t *testing.T) {
	service := &fakeRepositoryService{
		list: []RepositoryListItem{{ID: "1", Name: "demo"}},
	}
	useCase := NewRepoUseCase(service, &fakeGitRunner{})

	repos, err := useCase.ListRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected list to succeed, got: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != "demo" {
		t.Fatalf("unexpected repos: %+v", repos)
	}
}

func TestRepoUseCaseCloneUsesRepositoryDetail(t *testing.T) {
	service := &fakeRepositoryService{
		detail: RepositoryDetail{ID: "1", Name: "demo", CloneURL: "git@example.com:org/demo.git"},
	}
	git := &fakeGitRunner{}
	useCase := NewRepoUseCase(service, git)

	err := useCase.CloneRepository(context.Background(), "demo", "target-dir")
	if err != nil {
		t.Fatalf("expected clone to succeed, got: %v", err)
	}
	if !service.getCalled {
		t.Fatal("expected clone to request repository detail")
	}
	if git.cloneURL != "git@example.com:org/demo.git" {
		t.Fatalf("expected git clone URL from detail, got %q", git.cloneURL)
	}
	if git.destination != "target-dir" {
		t.Fatalf("expected destination target-dir, got %q", git.destination)
	}
}

func TestRepoUseCaseCloneRequiresCloneURL(t *testing.T) {
	service := &fakeRepositoryService{
		detail: RepositoryDetail{ID: "1", Name: "demo"},
	}
	git := &fakeGitRunner{}
	useCase := NewRepoUseCase(service, git)

	err := useCase.CloneRepository(context.Background(), "demo", "")
	if err == nil {
		t.Fatal("expected clone without clone URL to fail")
	}
	if git.called {
		t.Fatal("expected git not to be called without clone URL")
	}
}

type fakeRepositoryService struct {
	list      []RepositoryListItem
	detail    RepositoryDetail
	getCalled bool
	err       error
}

func (s *fakeRepositoryService) ListRepositories(ctx context.Context) ([]RepositoryListItem, error) {
	return s.list, s.err
}

func (s *fakeRepositoryService) GetRepository(ctx context.Context, id string) (RepositoryDetail, error) {
	s.getCalled = true
	if s.err != nil {
		return RepositoryDetail{}, s.err
	}
	if s.detail.ID == "" && s.detail.Name == "" {
		return RepositoryDetail{}, errors.New("not found")
	}
	return s.detail, nil
}

type fakeGitRunner struct {
	called      bool
	cloneURL    string
	destination string
	err         error
}

func (g *fakeGitRunner) Clone(ctx context.Context, cloneURL, destination string) error {
	g.called = true
	g.cloneURL = cloneURL
	g.destination = destination
	return g.err
}
