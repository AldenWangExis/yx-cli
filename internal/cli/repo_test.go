package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

func TestRepoListAndViewJSON(t *testing.T) {
	repos := &fakeRepoUseCase{
		list:     []app.RepositoryListItem{{ID: "1", Name: "demo", Path: "org/demo"}},
		detail:   app.RepositoryDetail{ID: "1", Name: "demo", Path: "org/demo", CloneURL: "git@example.com:org/demo.git"},
		created:  app.RepositoryMutationResult{Repository: app.RepositoryDetail{ID: "2", Name: "created", Path: "org/created"}},
		branches: []app.BranchListItem{{Name: "master", Default: true, CommitID: "abc123"}},
		commits:  []app.CommitListItem{{ID: "abc123", ShortID: "abc123", Title: "Initial commit"}},
		file:     app.RepositoryFile{Path: "test.py", Ref: "master", Content: "print(1)\n"},
	}
	opts := Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml"), RepoUseCase: repos}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "--json", "repo", "list")
	if err != nil {
		t.Fatalf("expected repo list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var listed []app.RepositoryListItem
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatalf("expected JSON repo list, got error: %v output=%s", err, stdout)
	}
	if len(listed) != 1 || listed[0].Name != "demo" {
		t.Fatalf("unexpected listed repos: %+v", listed)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "repo", "view", "demo")
	if err != nil {
		t.Fatalf("expected repo view to succeed, got error: %v stderr=%s", err, stderr)
	}
	var detail app.RepositoryDetail
	if err := json.Unmarshal([]byte(stdout), &detail); err != nil {
		t.Fatalf("expected JSON repo detail, got error: %v output=%s", err, stdout)
	}
	if detail.CloneURL != "git@example.com:org/demo.git" {
		t.Fatalf("expected clone URL in detail, got %q", detail.CloneURL)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "repo", "create", "--name", "created", "--path", "created", "--yes")
	if err != nil {
		t.Fatalf("expected repo create to succeed, got error: %v stderr=%s", err, stderr)
	}
	var created app.RepositoryMutationResult
	if err := json.Unmarshal([]byte(stdout), &created); err != nil {
		t.Fatalf("expected JSON repo create, got error: %v output=%s", err, stdout)
	}
	if repos.createInput.Name != "created" || repos.createInput.Path != "created" || !repos.createInput.Yes {
		t.Fatalf("unexpected create input: %+v", repos.createInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "repo", "branch", "list", "demo")
	if err != nil {
		t.Fatalf("expected branch list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var branches []app.BranchListItem
	if err := json.Unmarshal([]byte(stdout), &branches); err != nil {
		t.Fatalf("expected JSON branches, got error: %v output=%s", err, stdout)
	}
	if len(branches) != 1 || repos.branchRepo != "demo" {
		t.Fatalf("unexpected branches: %+v repo=%q", branches, repos.branchRepo)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "repo", "commit", "list", "demo", "--ref", "master")
	if err != nil {
		t.Fatalf("expected commit list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var commits []app.CommitListItem
	if err := json.Unmarshal([]byte(stdout), &commits); err != nil {
		t.Fatalf("expected JSON commits, got error: %v output=%s", err, stdout)
	}
	if len(commits) != 1 || repos.commitInput.Ref != "master" {
		t.Fatalf("unexpected commits: %+v input=%+v", commits, repos.commitInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "repo", "file", "view", "demo", "test.py", "--ref", "master")
	if err != nil {
		t.Fatalf("expected file view to succeed, got error: %v stderr=%s", err, stderr)
	}
	if stdout != "print(1)\n" {
		t.Fatalf("unexpected file output: %q", stdout)
	}
}

func TestRepoHelpShowsSubcommandsAndExamples(t *testing.T) {
	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:  filepath.Join(t.TempDir(), "config.yaml"),
		RepoUseCase: &fakeRepoUseCase{},
	}), "repo", "--help")
	if err != nil {
		t.Fatalf("expected repo help to succeed, got error: %v stderr=%s", err, stderr)
	}
	for _, want := range []string{
		"Manage Codeup repositories",
		"Available Commands:",
		"create",
		"branch",
		"commit",
		"file",
		"yx repo create --name demo --path demo --visibility private --yes",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected repo help to include %q, got:\n%s", want, stdout)
		}
	}
}

func TestRepoCreateHelpShowsFlags(t *testing.T) {
	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:  filepath.Join(t.TempDir(), "config.yaml"),
		RepoUseCase: &fakeRepoUseCase{},
	}), "repo", "create", "--help")
	if err != nil {
		t.Fatalf("expected repo create help to succeed, got error: %v stderr=%s", err, stderr)
	}
	for _, want := range []string{
		"Create a Codeup repository",
		"--name",
		"--path",
		"--visibility",
		"--dry-run",
		"--yes",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected repo create help to include %q, got:\n%s", want, stdout)
		}
	}
}

func TestRepoCloneCallsUseCase(t *testing.T) {
	repos := &fakeRepoUseCase{}
	opts := Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml"), RepoUseCase: repos}

	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "repo", "clone", "demo", "target-dir")
	if err != nil {
		t.Fatalf("expected repo clone to succeed, got error: %v stderr=%s", err, stderr)
	}
	if repos.cloneID != "demo" {
		t.Fatalf("expected clone id demo, got %q", repos.cloneID)
	}
	if repos.cloneDestination != "target-dir" {
		t.Fatalf("expected clone destination target-dir, got %q", repos.cloneDestination)
	}
}

func TestRepoViewRequiresRepositoryArgument(t *testing.T) {
	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:  filepath.Join(t.TempDir(), "config.yaml"),
		RepoUseCase: &fakeRepoUseCase{},
	}), "repo", "view")
	if err == nil {
		t.Fatal("expected repo view without repo to fail")
	}
	if stderr == "" {
		t.Fatal("expected argument error on stderr")
	}
}

type fakeRepoUseCase struct {
	list             []app.RepositoryListItem
	detail           app.RepositoryDetail
	created          app.RepositoryMutationResult
	branches         []app.BranchListItem
	commits          []app.CommitListItem
	file             app.RepositoryFile
	cloneID          string
	cloneDestination string
	createInput      app.CreateRepositoryInput
	branchRepo       string
	commitInput      app.CommitListInput
	fileInput        app.FileGetInput
	syncInput        app.BranchSyncInput
}

func (u *fakeRepoUseCase) ListRepositories(ctx context.Context) ([]app.RepositoryListItem, error) {
	return u.list, nil
}

func (u *fakeRepoUseCase) GetRepository(ctx context.Context, id string) (app.RepositoryDetail, error) {
	return u.detail, nil
}

func (u *fakeRepoUseCase) CreateRepository(ctx context.Context, input app.CreateRepositoryInput) (app.RepositoryMutationResult, error) {
	u.createInput = input
	return u.created, nil
}

func (u *fakeRepoUseCase) CloneRepository(ctx context.Context, id, destination string) error {
	u.cloneID = id
	u.cloneDestination = destination
	return nil
}

func (u *fakeRepoUseCase) ListBranches(ctx context.Context, repo string) ([]app.BranchListItem, error) {
	u.branchRepo = repo
	return u.branches, nil
}

func (u *fakeRepoUseCase) SyncBranch(ctx context.Context, input app.BranchSyncInput) (app.BranchMutationResult, error) {
	u.syncInput = input
	return app.BranchMutationResult{Branch: app.BranchListItem{Name: input.Target}}, nil
}

func (u *fakeRepoUseCase) ListCommits(ctx context.Context, input app.CommitListInput) ([]app.CommitListItem, error) {
	u.commitInput = input
	return u.commits, nil
}

func (u *fakeRepoUseCase) GetFile(ctx context.Context, input app.FileGetInput) (app.RepositoryFile, error) {
	u.fileInput = input
	return u.file, nil
}
