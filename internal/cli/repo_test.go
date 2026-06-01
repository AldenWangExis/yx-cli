package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

func TestRepoListAndViewJSON(t *testing.T) {
	repos := &fakeRepoUseCase{
		list:   []app.RepositoryListItem{{ID: "1", Name: "demo", Path: "org/demo"}},
		detail: app.RepositoryDetail{ID: "1", Name: "demo", Path: "org/demo", CloneURL: "git@example.com:org/demo.git"},
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
	cloneID          string
	cloneDestination string
}

func (u *fakeRepoUseCase) ListRepositories(ctx context.Context) ([]app.RepositoryListItem, error) {
	return u.list, nil
}

func (u *fakeRepoUseCase) GetRepository(ctx context.Context, id string) (app.RepositoryDetail, error) {
	return u.detail, nil
}

func (u *fakeRepoUseCase) CloneRepository(ctx context.Context, id, destination string) error {
	u.cloneID = id
	u.cloneDestination = destination
	return nil
}
