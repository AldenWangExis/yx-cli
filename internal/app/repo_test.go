package app

import (
	"context"
	"errors"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

func TestRepoUseCaseListsRepositories(t *testing.T) {
	service := &fakeRepositoryService{
		list: []RepositoryListItem{{ID: "1", Name: "demo"}},
	}
	useCase := NewRepoUseCase(service, &fakeGitRunner{}, safety.Environment{})

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
	useCase := NewRepoUseCase(service, git, safety.Environment{})

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
	useCase := NewRepoUseCase(service, git, safety.Environment{})

	err := useCase.CloneRepository(context.Background(), "demo", "")
	if err == nil {
		t.Fatal("expected clone without clone URL to fail")
	}
	if git.called {
		t.Fatal("expected git not to be called without clone URL")
	}
}

func TestRepoUseCaseCreatesRepositoryWithSafety(t *testing.T) {
	service := &fakeRepositoryService{
		created: RepositoryDetail{ID: "2", Name: "demo", Path: "org/demo"},
	}
	useCase := NewRepoUseCase(service, &fakeGitRunner{}, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	result, err := useCase.CreateRepository(context.Background(), CreateRepositoryInput{
		Name:       "demo",
		Path:       "demo",
		Visibility: "private",
		ReadmeType: "EMPTY",
		Yes:        true,
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got: %v", err)
	}
	if result.Repository.ID != "2" {
		t.Fatalf("unexpected repository: %+v", result.Repository)
	}
	if service.createInput.Visibility != "private" || service.createInput.ReadmeType != "EMPTY" {
		t.Fatalf("unexpected create input: %+v", service.createInput)
	}
}

func TestRepoUseCaseCreateRepositoryDryRunDoesNotMutate(t *testing.T) {
	service := &fakeRepositoryService{}
	useCase := NewRepoUseCase(service, &fakeGitRunner{}, safety.Environment{ConfirmWrites: true})

	result, err := useCase.CreateRepository(context.Background(), CreateRepositoryInput{Name: "demo", DryRun: true})
	if err != nil {
		t.Fatalf("expected dry-run to succeed, got: %v", err)
	}
	if !result.DryRun {
		t.Fatalf("expected dry-run result, got %+v", result)
	}
	if service.createCalled {
		t.Fatal("expected no create call during dry-run")
	}
}

func TestRepoUseCaseReadsRepositoryOperations(t *testing.T) {
	service := &fakeRepositoryService{
		branches: []BranchListItem{{Name: "master", Default: true}},
		commits:  []CommitListItem{{ID: "abc123", ShortID: "abc123", Title: "Initial commit"}},
		file:     RepositoryFile{Path: "test.py", Ref: "master", Content: "print(1)\n", Encoding: "text"},
	}
	useCase := NewRepoUseCase(service, &fakeGitRunner{}, safety.Environment{})

	branches, err := useCase.ListBranches(context.Background(), "repo-1")
	if err != nil {
		t.Fatalf("expected branches, got: %v", err)
	}
	if len(branches) != 1 || branches[0].Name != "master" {
		t.Fatalf("unexpected branches: %+v", branches)
	}

	commits, err := useCase.ListCommits(context.Background(), CommitListInput{Repo: "repo-1", Ref: "master"})
	if err != nil {
		t.Fatalf("expected commits, got: %v", err)
	}
	if len(commits) != 1 || commits[0].Title != "Initial commit" {
		t.Fatalf("unexpected commits: %+v", commits)
	}

	file, err := useCase.GetFile(context.Background(), FileGetInput{Repo: "repo-1", Path: "test.py", Ref: "master"})
	if err != nil {
		t.Fatalf("expected file, got: %v", err)
	}
	if file.Content != "print(1)\n" {
		t.Fatalf("unexpected file: %+v", file)
	}
}

func TestRepoUseCaseSyncBranchCreatesBranchFromSource(t *testing.T) {
	service := &fakeRepositoryService{
		synced: BranchListItem{Name: "feature/a", CommitID: "abc123"},
	}
	useCase := NewRepoUseCase(service, &fakeGitRunner{}, safety.Environment{})

	result, err := useCase.SyncBranch(context.Background(), BranchSyncInput{
		Repo:   "repo-1",
		Source: "master",
		Target: "feature/a",
		Yes:    true,
	})
	if err != nil {
		t.Fatalf("expected branch sync, got: %v", err)
	}
	if result.Branch.Name != "feature/a" {
		t.Fatalf("unexpected branch result: %+v", result)
	}
	if service.syncInput.Source != "master" || service.syncInput.Target != "feature/a" {
		t.Fatalf("unexpected sync input: %+v", service.syncInput)
	}
}

func TestRepoUseCaseDeleteRepositoryHonorsSafety(t *testing.T) {
	service := &fakeRepositoryService{}
	useCase := NewRepoUseCase(service, &fakeGitRunner{}, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	result, err := useCase.DeleteRepository(context.Background(), DeleteRepositoryInput{Repo: "repo-1", DryRun: true})
	if err != nil {
		t.Fatalf("expected delete dry-run to succeed, got: %v", err)
	}
	if !result.DryRun || result.Summary != "delete repository repo-1" {
		t.Fatalf("unexpected dry-run result: %+v", result)
	}
	if service.deleteCalled {
		t.Fatal("expected dry-run not to call delete")
	}

	_, err = useCase.DeleteRepository(context.Background(), DeleteRepositoryInput{Repo: "repo-1"})
	if err == nil {
		t.Fatal("expected delete without yes or dry-run to require confirmation")
	}
	if service.deleteCalled {
		t.Fatal("expected delete not to be called without confirmation")
	}

	result, err = useCase.DeleteRepository(context.Background(), DeleteRepositoryInput{Repo: "repo-1", Yes: true})
	if err != nil {
		t.Fatalf("expected delete with yes to succeed, got: %v", err)
	}
	if result.Repository.ID != "repo-1" || service.deletedRepo != "repo-1" {
		t.Fatalf("unexpected delete result=%+v deleted=%q", result, service.deletedRepo)
	}
}

func TestRepoUseCaseRepositoryMemberOperationsHonorSafety(t *testing.T) {
	service := &fakeRepositoryService{
		members: []RepositoryMember{{UserID: "u1", Name: "A", AccessLevel: 30, Access: "developer"}},
		member:  RepositoryMember{UserID: "u1", AccessLevel: 40, Access: "maintainer"},
	}
	useCase := NewRepoUseCase(service, &fakeGitRunner{}, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	members, err := useCase.ListRepositoryMembers(context.Background(), "repo-1")
	if err != nil {
		t.Fatalf("expected member list, got: %v", err)
	}
	if len(members) != 1 || members[0].Access != "developer" {
		t.Fatalf("unexpected members: %+v", members)
	}

	added, err := useCase.AddRepositoryMember(context.Background(), AddRepositoryMemberInput{
		Repo:        "repo-1",
		UserID:      "u1",
		AccessLevel: "developer",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("expected add dry-run, got: %v", err)
	}
	if !added.DryRun || service.addCalled {
		t.Fatalf("expected dry-run without mutation, result=%+v called=%v", added, service.addCalled)
	}

	_, err = useCase.UpdateRepositoryMember(context.Background(), UpdateRepositoryMemberInput{
		Repo:        "repo-1",
		UserID:      "u1",
		AccessLevel: "maintainer",
	})
	if err == nil {
		t.Fatal("expected update without yes or dry-run to require confirmation")
	}
	if service.updateCalled {
		t.Fatal("expected no update without confirmation")
	}

	updated, err := useCase.UpdateRepositoryMember(context.Background(), UpdateRepositoryMemberInput{
		Repo:        "repo-1",
		UserID:      "u1",
		AccessLevel: "maintainer",
		Yes:         true,
	})
	if err != nil {
		t.Fatalf("expected update with yes, got: %v", err)
	}
	if updated.Member.Access != "maintainer" || service.updateInput.AccessLevel != "40" {
		t.Fatalf("unexpected update result=%+v input=%+v", updated, service.updateInput)
	}

	removed, err := useCase.RemoveRepositoryMember(context.Background(), RemoveRepositoryMemberInput{
		Repo:   "repo-1",
		UserID: "u1",
		Yes:    true,
	})
	if err != nil {
		t.Fatalf("expected remove with yes, got: %v", err)
	}
	if removed.Member.UserID != "u1" || service.removeInput.UserID != "u1" {
		t.Fatalf("unexpected remove result=%+v input=%+v", removed, service.removeInput)
	}
}

type fakeRepositoryService struct {
	list         []RepositoryListItem
	detail       RepositoryDetail
	created      RepositoryDetail
	branches     []BranchListItem
	commits      []CommitListItem
	file         RepositoryFile
	members      []RepositoryMember
	member       RepositoryMember
	synced       BranchListItem
	createInput  CreateRepositoryInput
	syncInput    BranchSyncInput
	addInput     AddRepositoryMemberInput
	updateInput  UpdateRepositoryMemberInput
	removeInput  RemoveRepositoryMemberInput
	listCalls    int
	getCalled    bool
	createCalled bool
	addCalled    bool
	updateCalled bool
	removeCalled bool
	deleteCalled bool
	deletedRepo  string
	err          error
}

func (s *fakeRepositoryService) ListRepositories(ctx context.Context) ([]RepositoryListItem, error) {
	s.listCalls++
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

func (s *fakeRepositoryService) CreateRepository(ctx context.Context, input CreateRepositoryInput) (RepositoryDetail, error) {
	s.createCalled = true
	s.createInput = input
	return s.created, s.err
}

func (s *fakeRepositoryService) ListBranches(ctx context.Context, repo string) ([]BranchListItem, error) {
	return s.branches, s.err
}

func (s *fakeRepositoryService) ListCommits(ctx context.Context, input CommitListInput) ([]CommitListItem, error) {
	return s.commits, s.err
}

func (s *fakeRepositoryService) GetFile(ctx context.Context, input FileGetInput) (RepositoryFile, error) {
	return s.file, s.err
}

func (s *fakeRepositoryService) SyncBranch(ctx context.Context, input BranchSyncInput) (BranchListItem, error) {
	s.syncInput = input
	return s.synced, s.err
}

func (s *fakeRepositoryService) DeleteRepository(ctx context.Context, repo string) (RepositoryDetail, error) {
	s.deleteCalled = true
	s.deletedRepo = repo
	return RepositoryDetail{ID: repo}, s.err
}

func (s *fakeRepositoryService) ListRepositoryMembers(ctx context.Context, repo string) ([]RepositoryMember, error) {
	return s.members, s.err
}

func (s *fakeRepositoryService) AddRepositoryMember(ctx context.Context, input AddRepositoryMemberInput) (RepositoryMember, error) {
	s.addCalled = true
	s.addInput = input
	return s.member, s.err
}

func (s *fakeRepositoryService) UpdateRepositoryMember(ctx context.Context, input UpdateRepositoryMemberInput) (RepositoryMember, error) {
	s.updateCalled = true
	s.updateInput = input
	return s.member, s.err
}

func (s *fakeRepositoryService) RemoveRepositoryMember(ctx context.Context, input RemoveRepositoryMemberInput) (RepositoryMember, error) {
	s.removeCalled = true
	s.removeInput = input
	return RepositoryMember{UserID: input.UserID}, s.err
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
