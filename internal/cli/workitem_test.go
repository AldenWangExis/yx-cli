package cli

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

func TestProjectAndWorkitemCommands(t *testing.T) {
	workitems := &fakeWorkitemUseCase{
		projects: []app.Project{{ID: "p1", Name: "Project One"}},
		list:     []app.WorkitemListItem{{ID: "w1", Title: "Task One", Status: "todo", Type: "task", ProjectID: "p1"}},
		detail:   app.WorkitemDetail{ID: "w1", Title: "Task One", Status: "todo", Type: "task", ProjectID: "p1"},
		created:  app.WorkitemMutationResult{DryRun: true, Summary: "create task"},
		updated:  app.WorkitemMutationResult{DryRun: true, Summary: "update task"},
	}
	opts := Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml"), WorkitemUseCase: workitems}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "--json", "project", "list")
	if err != nil {
		t.Fatalf("expected project list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var projects []app.Project
	if err := json.Unmarshal([]byte(stdout), &projects); err != nil {
		t.Fatalf("expected project JSON, got error: %v output=%s", err, stdout)
	}
	if len(projects) != 1 || projects[0].ID != "p1" {
		t.Fatalf("unexpected projects: %+v", projects)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "project", "create", "--name", "测试", "--yes")
	if err != nil {
		t.Fatalf("expected project create to succeed, got error: %v stderr=%s", err, stderr)
	}
	var created app.ProjectMutationResult
	if err := json.Unmarshal([]byte(stdout), &created); err != nil {
		t.Fatalf("expected project create JSON, got error: %v output=%s", err, stdout)
	}
	if workitems.projectCreateInput.Name != "测试" || workitems.projectCreateInput.Scope != "public" || !workitems.projectCreateInput.Yes {
		t.Fatalf("unexpected project create input: %+v", workitems.projectCreateInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "workitem", "list", "--project", "p1")
	if err != nil {
		t.Fatalf("expected workitem list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var items []app.WorkitemListItem
	if err := json.Unmarshal([]byte(stdout), &items); err != nil {
		t.Fatalf("expected workitem JSON, got error: %v output=%s", err, stdout)
	}
	if len(items) != 1 || workitems.listInput.ProjectID != "p1" {
		t.Fatalf("unexpected workitems: %+v input=%+v", items, workitems.listInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "workitem", "view", "w1")
	if err != nil {
		t.Fatalf("expected workitem view to succeed, got error: %v stderr=%s", err, stderr)
	}
	var detail app.WorkitemDetail
	if err := json.Unmarshal([]byte(stdout), &detail); err != nil {
		t.Fatalf("expected workitem detail JSON, got error: %v output=%s", err, stdout)
	}
	if detail.ID != "w1" {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts),
		"workitem", "create", "--project", "p1", "--type", "task", "--title", "Task One", "--description", "Details", "--description-format", "markdown", "--dry-run")
	if err != nil {
		t.Fatalf("expected workitem create to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !workitems.createInput.DryRun ||
		workitems.createInput.ProjectID != "p1" ||
		workitems.createInput.Description != "Details" ||
		workitems.createInput.DescriptionFormat != "markdown" {
		t.Fatalf("unexpected create input: %+v", workitems.createInput)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts),
		"workitem", "update", "w1", "--status", "done", "--assignee", "u1", "--dry-run")
	if err != nil {
		t.Fatalf("expected workitem update to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !workitems.updateInput.DryRun || workitems.updateInput.ID != "w1" || workitems.updateInput.Status != "done" {
		t.Fatalf("unexpected update input: %+v", workitems.updateInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts),
		"issue", "delete", "w1", "--dry-run")
	if err != nil {
		t.Fatalf("expected issue delete dry-run to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "dry-run: delete workitem w1") {
		t.Fatalf("expected dry-run delete output, got:\n%s", stdout)
	}
	if !workitems.deleteInput.DryRun || workitems.deleteInput.ID != "w1" {
		t.Fatalf("unexpected delete input: %+v", workitems.deleteInput)
	}
}

func TestIssueAliasUsesWorkitemUseCase(t *testing.T) {
	workitems := &fakeWorkitemUseCase{
		list: []app.WorkitemListItem{{ID: "w1", Title: "Task One", ProjectID: "p1"}},
	}
	opts := Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml"), WorkitemUseCase: workitems}

	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "issue", "list", "--repo", "repo-a")
	if err != nil {
		t.Fatalf("expected issue list to succeed, got error: %v stderr=%s", err, stderr)
	}
	if workitems.listInput.Repo != "repo-a" {
		t.Fatalf("expected issue alias to pass repo, got %+v", workitems.listInput)
	}
}

func TestIssueListDefaultsToCurrentRepositoryMapping(t *testing.T) {
	workitems := &fakeWorkitemUseCase{
		list: []app.WorkitemListItem{{ID: "w1", Title: "Task One", ProjectID: "p1"}},
	}
	resolver := &fakeRepoCurrentResolver{
		current: app.CurrentRepository{ID: "6925918", Name: "yx-cli", Path: "org/yx-cli"},
	}
	opts := Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		WorkitemUseCase:     workitems,
		RepoCurrentResolver: resolver,
	}

	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "issue", "list")
	if err != nil {
		t.Fatalf("expected issue list to default to current repo, got error: %v stderr=%s", err, stderr)
	}
	if workitems.listInput.Repo != "6925918" {
		t.Fatalf("expected issue list to use current repo id, got %+v", workitems.listInput)
	}
	if resolver.calls != 1 {
		t.Fatalf("expected current repo resolver to be called once, got %d", resolver.calls)
	}
}

func TestIssueListExplicitProjectDoesNotResolveCurrentRepository(t *testing.T) {
	workitems := &fakeWorkitemUseCase{}
	resolver := &fakeRepoCurrentResolver{
		current: app.CurrentRepository{ID: "current"},
	}
	opts := Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		WorkitemUseCase:     workitems,
		RepoCurrentResolver: resolver,
	}

	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "issue", "list", "--project", "p1")
	if err != nil {
		t.Fatalf("expected issue list with explicit project to succeed, got error: %v stderr=%s", err, stderr)
	}
	if workitems.listInput.ProjectID != "p1" {
		t.Fatalf("expected explicit project, got %+v", workitems.listInput)
	}
	if resolver.calls != 0 {
		t.Fatalf("expected explicit project not to call current resolver, got %d", resolver.calls)
	}
}

func TestIssueRepoMappingErrorIsStable(t *testing.T) {
	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:      filepath.Join(t.TempDir(), "config.yaml"),
		WorkitemUseCase: &failingMappingWorkitemUseCase{},
	}), "issue", "list", "--repo", "repo-a")
	if err == nil {
		t.Fatal("expected missing mapping to fail")
	}
	if !strings.Contains(stderr, `repo "repo-a" is not mapped to a project`) {
		t.Fatalf("expected stable mapping error, got:\n%s", stderr)
	}
}

var errFakeCurrentRepository = errors.New("current repository is unavailable")

type fakeWorkitemUseCase struct {
	projects           []app.Project
	list               []app.WorkitemListItem
	detail             app.WorkitemDetail
	created            app.WorkitemMutationResult
	projectCreated     app.ProjectMutationResult
	updated            app.WorkitemMutationResult
	deleted            app.WorkitemMutationResult
	listInput          app.WorkitemListInput
	projectCreateInput app.CreateProjectInput
	createInput        app.CreateWorkitemInput
	updateInput        app.UpdateWorkitemInput
	deleteInput        app.DeleteWorkitemInput
}

func (u *fakeWorkitemUseCase) ListProjects(ctx context.Context) ([]app.Project, error) {
	return u.projects, nil
}

func (u *fakeWorkitemUseCase) CreateProject(ctx context.Context, input app.CreateProjectInput) (app.ProjectMutationResult, error) {
	u.projectCreateInput = input
	if u.projectCreated.Project.ID == "" {
		return app.ProjectMutationResult{Project: app.Project{ID: "p2", Name: input.Name}}, nil
	}
	return u.projectCreated, nil
}

func (u *fakeWorkitemUseCase) ListWorkitems(ctx context.Context, input app.WorkitemListInput) ([]app.WorkitemListItem, error) {
	u.listInput = input
	return u.list, nil
}

func (u *fakeWorkitemUseCase) GetWorkitem(ctx context.Context, id string) (app.WorkitemDetail, error) {
	return u.detail, nil
}

func (u *fakeWorkitemUseCase) CreateWorkitem(ctx context.Context, input app.CreateWorkitemInput) (app.WorkitemMutationResult, error) {
	u.createInput = input
	return u.created, nil
}

func (u *fakeWorkitemUseCase) UpdateWorkitem(ctx context.Context, input app.UpdateWorkitemInput) (app.WorkitemMutationResult, error) {
	u.updateInput = input
	return u.updated, nil
}

func (u *fakeWorkitemUseCase) DeleteWorkitem(ctx context.Context, input app.DeleteWorkitemInput) (app.WorkitemMutationResult, error) {
	u.deleteInput = input
	if u.deleted.Summary == "" {
		return app.WorkitemMutationResult{DryRun: input.DryRun, Summary: "delete workitem " + input.ID}, nil
	}
	return u.deleted, nil
}

type failingMappingWorkitemUseCase struct{}

func (u *failingMappingWorkitemUseCase) ListProjects(ctx context.Context) ([]app.Project, error) {
	return nil, nil
}

func (u *failingMappingWorkitemUseCase) CreateProject(ctx context.Context, input app.CreateProjectInput) (app.ProjectMutationResult, error) {
	return app.ProjectMutationResult{}, nil
}

func (u *failingMappingWorkitemUseCase) ListWorkitems(ctx context.Context, input app.WorkitemListInput) ([]app.WorkitemListItem, error) {
	return nil, app.ErrMissingRepoProjectMapping{Repo: input.Repo}
}

func (u *failingMappingWorkitemUseCase) GetWorkitem(ctx context.Context, id string) (app.WorkitemDetail, error) {
	return app.WorkitemDetail{}, nil
}

func (u *failingMappingWorkitemUseCase) CreateWorkitem(ctx context.Context, input app.CreateWorkitemInput) (app.WorkitemMutationResult, error) {
	return app.WorkitemMutationResult{}, nil
}

func (u *failingMappingWorkitemUseCase) UpdateWorkitem(ctx context.Context, input app.UpdateWorkitemInput) (app.WorkitemMutationResult, error) {
	return app.WorkitemMutationResult{}, nil
}

func (u *failingMappingWorkitemUseCase) DeleteWorkitem(ctx context.Context, input app.DeleteWorkitemInput) (app.WorkitemMutationResult, error) {
	return app.WorkitemMutationResult{}, nil
}
