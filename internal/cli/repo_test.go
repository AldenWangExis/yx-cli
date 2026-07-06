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

func TestRepoCurrentJSONCallsResolver(t *testing.T) {
	resolver := &fakeRepoCurrentResolver{
		current: app.CurrentRepository{
			ID:        "6925918",
			Name:      "yx-cli",
			Path:      "68086322e3a71588779435e0/yx-cli",
			Remote:    "codeup",
			RemoteURL: "git@codeup.aliyun.com:68086322e3a71588779435e0/yx-cli.git",
			Source:    "api",
		},
	}
	opts := Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		RepoCurrentResolver: resolver,
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts),
		"--json", "repo", "current", "--remote", "codeup", "--refresh")
	if err != nil {
		t.Fatalf("expected repo current to succeed, got error: %v stderr=%s", err, stderr)
	}
	var current app.CurrentRepository
	if err := json.Unmarshal([]byte(stdout), &current); err != nil {
		t.Fatalf("expected JSON current repo, got error: %v output=%s", err, stdout)
	}
	if current.ID != "6925918" || current.Remote != "codeup" {
		t.Fatalf("unexpected current repo: %+v", current)
	}
	if resolver.input.Remote != "codeup" || !resolver.input.Refresh {
		t.Fatalf("unexpected resolver input: %+v", resolver.input)
	}
	if resolver.input.WorkDir == "" {
		t.Fatal("expected repo current to pass a working directory")
	}
}

func TestRepoReadCommandsDefaultToCurrentRepository(t *testing.T) {
	repos := &fakeRepoUseCase{
		detail:   app.RepositoryDetail{ID: "6925918", Name: "yx-cli", Path: "org/yx-cli"},
		branches: []app.BranchListItem{{Name: "master", Default: true}},
		commits:  []app.CommitListItem{{ID: "abc123", ShortID: "abc123", Title: "Initial commit"}},
		file:     app.RepositoryFile{Path: "test.py", Ref: "master", Content: "print(1)\n"},
	}
	resolver := &fakeRepoCurrentResolver{
		current: app.CurrentRepository{ID: "6925918", Name: "yx-cli", Path: "org/yx-cli"},
	}
	opts := Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		RepoUseCase:         repos,
		RepoCurrentResolver: resolver,
	}

	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "repo", "view")
	if err != nil {
		t.Fatalf("expected repo view to default to current repo, got error: %v stderr=%s", err, stderr)
	}
	if repos.getID != "6925918" {
		t.Fatalf("expected repo view to use current repo id, got %q", repos.getID)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "repo", "branch", "list")
	if err != nil {
		t.Fatalf("expected branch list to default to current repo, got error: %v stderr=%s", err, stderr)
	}
	if repos.branchRepo != "6925918" {
		t.Fatalf("expected branch list to use current repo id, got %q", repos.branchRepo)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "repo", "commit", "list", "--ref", "master")
	if err != nil {
		t.Fatalf("expected commit list to default to current repo, got error: %v stderr=%s", err, stderr)
	}
	if repos.commitInput.Repo != "6925918" {
		t.Fatalf("expected commit list to use current repo id, got %+v", repos.commitInput)
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "repo", "file", "view", "test.py", "--ref", "master")
	if err != nil {
		t.Fatalf("expected file view to default to current repo, got error: %v stderr=%s", err, stderr)
	}
	if stdout != "print(1)\n" || repos.fileInput.Repo != "6925918" || repos.fileInput.Path != "test.py" {
		t.Fatalf("unexpected file default output=%q input=%+v", stdout, repos.fileInput)
	}

	if resolver.calls != 4 {
		t.Fatalf("expected current repo resolver to be called four times, got %d", resolver.calls)
	}
}

func TestRepoBranchSyncDefaultsToCurrentRepository(t *testing.T) {
	repos := &fakeRepoUseCase{}
	resolver := &fakeRepoCurrentResolver{
		current: app.CurrentRepository{ID: "6925918", Name: "yx-cli", Path: "org/yx-cli"},
	}
	opts := Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		RepoUseCase:         repos,
		RepoCurrentResolver: resolver,
	}

	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts),
		"repo", "branch", "sync", "--source", "master", "--target", "feat/a", "--dry-run")
	if err != nil {
		t.Fatalf("expected branch sync to default to current repo, got error: %v stderr=%s", err, stderr)
	}
	if repos.syncInput.Repo != "6925918" || repos.syncInput.Source != "master" || repos.syncInput.Target != "feat/a" {
		t.Fatalf("expected branch sync to use current repo id, got %+v", repos.syncInput)
	}
}

func TestRepoExplicitArgumentDoesNotResolveCurrentRepository(t *testing.T) {
	repos := &fakeRepoUseCase{
		branches: []app.BranchListItem{{Name: "master"}},
	}
	resolver := &fakeRepoCurrentResolver{
		current: app.CurrentRepository{ID: "current"},
	}
	opts := Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		RepoUseCase:         repos,
		RepoCurrentResolver: resolver,
	}

	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "repo", "branch", "list", "explicit")
	if err != nil {
		t.Fatalf("expected explicit branch repo to succeed, got error: %v stderr=%s", err, stderr)
	}
	if repos.branchRepo != "explicit" {
		t.Fatalf("expected explicit branch repo, got %q", repos.branchRepo)
	}
	if resolver.calls != 0 {
		t.Fatalf("expected explicit repo not to call current resolver, got %d", resolver.calls)
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
		"current",
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

func TestRepoDeletePassesSafetyFlagsAndDefaultsToCurrentRepository(t *testing.T) {
	repos := &fakeRepoUseCase{
		deleted: app.RepositoryMutationResult{DryRun: true, Summary: "delete repository 6925918"},
	}
	resolver := &fakeRepoCurrentResolver{
		current: app.CurrentRepository{ID: "6925918", Name: "yx-cli", Path: "org/yx-cli"},
	}
	opts := Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		RepoUseCase:         repos,
		RepoCurrentResolver: resolver,
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "repo", "delete", "--dry-run")
	if err != nil {
		t.Fatalf("expected repo delete dry-run to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "dry-run: delete repository 6925918") {
		t.Fatalf("expected dry-run output, got:\n%s", stdout)
	}
	if repos.deleteInput.Repo != "6925918" || !repos.deleteInput.DryRun || repos.deleteInput.Yes {
		t.Fatalf("unexpected delete input: %+v", repos.deleteInput)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "repo", "delete", "explicit", "--yes")
	if err != nil {
		t.Fatalf("expected repo delete --yes to succeed, got error: %v stderr=%s", err, stderr)
	}
	if repos.deleteInput.Repo != "explicit" || !repos.deleteInput.Yes {
		t.Fatalf("unexpected explicit delete input: %+v", repos.deleteInput)
	}
}

func TestRepoMemberCommands(t *testing.T) {
	repos := &fakeRepoUseCase{
		members: []app.RepositoryMember{{UserID: "u1", Name: "A", Email: "a@example.com", AccessLevel: 30, Access: "developer"}},
		memberMutation: app.RepositoryMemberMutationResult{
			DryRun:  true,
			Summary: "add repository member u1 to repo-1 as 30",
		},
	}
	resolver := &fakeRepoCurrentResolver{
		current: app.CurrentRepository{ID: "repo-1", Name: "yx-cli", Path: "org/yx-cli"},
	}
	opts := Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		RepoUseCase:         repos,
		RepoCurrentResolver: resolver,
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "--json", "repo", "member", "list")
	if err != nil {
		t.Fatalf("expected repo member list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var members []app.RepositoryMember
	if err := json.Unmarshal([]byte(stdout), &members); err != nil {
		t.Fatalf("expected JSON members, got error: %v output=%s", err, stdout)
	}
	if len(members) != 1 || repos.memberRepo != "repo-1" {
		t.Fatalf("unexpected members=%+v repo=%q", members, repos.memberRepo)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "repo", "member", "add", "--user-id", "u1", "--access-level", "developer", "--dry-run")
	if err != nil {
		t.Fatalf("expected repo member add to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "dry-run: add repository member u1 to repo-1 as 30") {
		t.Fatalf("expected dry-run output, got:\n%s", stdout)
	}
	if repos.addMemberInput.Repo != "repo-1" || repos.addMemberInput.UserID != "u1" || repos.addMemberInput.AccessLevel != "developer" || !repos.addMemberInput.DryRun {
		t.Fatalf("unexpected add input: %+v", repos.addMemberInput)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "repo", "member", "update", "explicit", "--user-id", "u1", "--access-level", "maintainer", "--yes")
	if err != nil {
		t.Fatalf("expected repo member update to succeed, got error: %v stderr=%s", err, stderr)
	}
	if repos.updateMemberInput.Repo != "explicit" || repos.updateMemberInput.AccessLevel != "maintainer" || !repos.updateMemberInput.Yes {
		t.Fatalf("unexpected update input: %+v", repos.updateMemberInput)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "repo", "member", "remove", "explicit", "--user-id", "u1", "--yes")
	if err != nil {
		t.Fatalf("expected repo member remove to succeed, got error: %v stderr=%s", err, stderr)
	}
	if repos.removeMemberInput.Repo != "explicit" || repos.removeMemberInput.UserID != "u1" || !repos.removeMemberInput.Yes {
		t.Fatalf("unexpected remove input: %+v", repos.removeMemberInput)
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
	list              []app.RepositoryListItem
	detail            app.RepositoryDetail
	created           app.RepositoryMutationResult
	branches          []app.BranchListItem
	commits           []app.CommitListItem
	file              app.RepositoryFile
	deleted           app.RepositoryMutationResult
	members           []app.RepositoryMember
	memberMutation    app.RepositoryMemberMutationResult
	cloneID           string
	cloneDestination  string
	createInput       app.CreateRepositoryInput
	getID             string
	branchRepo        string
	commitInput       app.CommitListInput
	fileInput         app.FileGetInput
	syncInput         app.BranchSyncInput
	deleteInput       app.DeleteRepositoryInput
	memberRepo        string
	addMemberInput    app.AddRepositoryMemberInput
	updateMemberInput app.UpdateRepositoryMemberInput
	removeMemberInput app.RemoveRepositoryMemberInput
}

type fakeRepoCurrentResolver struct {
	current app.CurrentRepository
	input   app.CurrentRepositoryInput
	calls   int
	err     error
}

func (r *fakeRepoCurrentResolver) CurrentRepository(ctx context.Context, input app.CurrentRepositoryInput) (app.CurrentRepository, error) {
	r.calls++
	r.input = input
	return r.current, r.err
}

func (u *fakeRepoUseCase) ListRepositories(ctx context.Context) ([]app.RepositoryListItem, error) {
	return u.list, nil
}

func (u *fakeRepoUseCase) GetRepository(ctx context.Context, id string) (app.RepositoryDetail, error) {
	u.getID = id
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

func (u *fakeRepoUseCase) DeleteRepository(ctx context.Context, input app.DeleteRepositoryInput) (app.RepositoryMutationResult, error) {
	u.deleteInput = input
	return u.deleted, nil
}

func (u *fakeRepoUseCase) ListRepositoryMembers(ctx context.Context, repo string) ([]app.RepositoryMember, error) {
	u.memberRepo = repo
	return u.members, nil
}

func (u *fakeRepoUseCase) AddRepositoryMember(ctx context.Context, input app.AddRepositoryMemberInput) (app.RepositoryMemberMutationResult, error) {
	u.addMemberInput = input
	if u.memberMutation.DryRun || u.memberMutation.Member.UserID != "" {
		return u.memberMutation, nil
	}
	return app.RepositoryMemberMutationResult{Member: app.RepositoryMember{UserID: input.UserID, Access: input.AccessLevel}}, nil
}

func (u *fakeRepoUseCase) UpdateRepositoryMember(ctx context.Context, input app.UpdateRepositoryMemberInput) (app.RepositoryMemberMutationResult, error) {
	u.updateMemberInput = input
	return app.RepositoryMemberMutationResult{Member: app.RepositoryMember{UserID: input.UserID, Access: input.AccessLevel}}, nil
}

func (u *fakeRepoUseCase) RemoveRepositoryMember(ctx context.Context, input app.RemoveRepositoryMemberInput) (app.RepositoryMemberMutationResult, error) {
	u.removeMemberInput = input
	return app.RepositoryMemberMutationResult{Member: app.RepositoryMember{UserID: input.UserID}}, nil
}
