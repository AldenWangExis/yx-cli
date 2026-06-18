package app

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

func TestWorkitemUseCaseProjectFirst(t *testing.T) {
	projects := &fakeProjectService{projects: []Project{{ID: "p1", Name: "Project One"}}}
	workitems := &fakeWorkitemService{
		list:   []WorkitemListItem{{ID: "w1", Title: "Task One", Status: "todo", Type: "task", ProjectID: "p1"}},
		detail: WorkitemDetail{ID: "w1", Title: "Task One", Status: "todo", Type: "task", ProjectID: "p1"},
	}
	useCase := NewWorkitemUseCase(projects, workitems, map[string]string{}, safety.Environment{})

	listedProjects, err := useCase.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("expected projects to list, got: %v", err)
	}
	if len(listedProjects) != 1 || listedProjects[0].ID != "p1" {
		t.Fatalf("unexpected projects: %+v", listedProjects)
	}

	items, err := useCase.ListWorkitems(context.Background(), WorkitemListInput{ProjectID: "p1"})
	if err != nil {
		t.Fatalf("expected workitems to list, got: %v", err)
	}
	if len(items) != 1 || items[0].ProjectID != "p1" {
		t.Fatalf("unexpected workitems: %+v", items)
	}

	detail, err := useCase.GetWorkitem(context.Background(), "w1")
	if err != nil {
		t.Fatalf("expected workitem detail, got: %v", err)
	}
	if detail.Title != "Task One" {
		t.Fatalf("unexpected detail: %+v", detail)
	}
}

func TestWorkitemUseCaseRepoMapping(t *testing.T) {
	workitems := &fakeWorkitemService{
		list: []WorkitemListItem{{ID: "w1", Title: "Task One", ProjectID: "p1"}},
	}
	useCase := NewWorkitemUseCase(&fakeProjectService{}, workitems, map[string]string{"repo-a": "p1"}, safety.Environment{})

	items, err := useCase.ListWorkitems(context.Background(), WorkitemListInput{Repo: "repo-a"})
	if err != nil {
		t.Fatalf("expected repo mapping to resolve, got: %v", err)
	}
	if len(items) != 1 || workitems.lastProjectID != "p1" {
		t.Fatalf("expected project p1, got items=%+v project=%q", items, workitems.lastProjectID)
	}
}

func TestWorkitemUseCaseMissingRepoMappingDoesNotCallService(t *testing.T) {
	workitems := &fakeWorkitemService{}
	useCase := NewWorkitemUseCase(&fakeProjectService{}, workitems, map[string]string{}, safety.Environment{})

	_, err := useCase.ListWorkitems(context.Background(), WorkitemListInput{Repo: "repo-a"})
	if err == nil {
		t.Fatal("expected missing mapping to fail")
	}
	if !strings.Contains(err.Error(), "profiles.<profile>.repoProjectMap.repo-a") {
		t.Fatalf("expected actionable repo project mapping command, got: %v", err)
	}
	if workitems.listCalled {
		t.Fatal("expected service not to be called without mapping")
	}
}

func TestProjectCreateUsesDefaultTemplateAndGeneratedCode(t *testing.T) {
	projects := &fakeProjectService{
		templates: []ProjectTemplate{{ID: "tpl-1", Name: "Classic"}},
		created:   Project{ID: "p2"},
	}
	useCase := NewWorkitemUseCase(projects, &fakeWorkitemService{}, nil, safety.Environment{})

	result, err := useCase.CreateProject(context.Background(), CreateProjectInput{Name: "测试"})
	if err != nil {
		t.Fatalf("expected project create to succeed, got: %v", err)
	}
	if result.DryRun {
		t.Fatalf("expected real create, got dry-run: %+v", result)
	}
	if result.Project.ID != "p2" {
		t.Fatalf("unexpected project: %+v", result.Project)
	}
	if result.Project.Name != "测试" {
		t.Fatalf("expected name to be filled from input, got %+v", result.Project)
	}
	if projects.createInput.TemplateID != "tpl-1" {
		t.Fatalf("expected default template, got %+v", projects.createInput)
	}
	if projects.createInput.Scope != "public" {
		t.Fatalf("expected default public scope, got %+v", projects.createInput)
	}
	if len(projects.createInput.CustomCode) < 4 || len(projects.createInput.CustomCode) > 6 {
		t.Fatalf("expected generated custom code, got %+v", projects.createInput)
	}
}

func TestProjectCreateDryRunDoesNotMutate(t *testing.T) {
	projects := &fakeProjectService{templates: []ProjectTemplate{{ID: "tpl-1", Name: "Classic"}}}
	useCase := NewWorkitemUseCase(projects, &fakeWorkitemService{}, nil, safety.Environment{ConfirmWrites: true})

	result, err := useCase.CreateProject(context.Background(), CreateProjectInput{Name: "测试", DryRun: true})
	if err != nil {
		t.Fatalf("expected dry-run to succeed, got: %v", err)
	}
	if !result.DryRun {
		t.Fatalf("expected dry-run result, got %+v", result)
	}
	if projects.createCalled {
		t.Fatal("expected no mutation during dry-run")
	}
}

func TestWorkitemCreateAndUpdateDryRunDoNotMutate(t *testing.T) {
	workitems := &fakeWorkitemService{}
	useCase := NewWorkitemUseCase(&fakeProjectService{}, workitems, nil, safety.Environment{ConfirmWrites: true})

	created, err := useCase.CreateWorkitem(context.Background(), CreateWorkitemInput{
		ProjectID:         "p1",
		Type:              "task",
		Title:             "Task One",
		Description:       "Details",
		DescriptionFormat: "markdown",
		Assignee:          "u1",
		DryRun:            true,
	})
	if err != nil {
		t.Fatalf("expected create dry-run to succeed, got: %v", err)
	}
	if !created.DryRun || workitems.createCalled {
		t.Fatalf("expected dry-run without mutation, result=%+v called=%v", created, workitems.createCalled)
	}

	_, err = useCase.CreateWorkitem(context.Background(), CreateWorkitemInput{
		ProjectID:         "p1",
		Type:              "task",
		Title:             "Task One",
		DescriptionFormat: "html",
	})
	if err == nil || !strings.Contains(err.Error(), "description format") {
		t.Fatalf("expected invalid description format error, got: %v", err)
	}

	updated, err := useCase.UpdateWorkitem(context.Background(), UpdateWorkitemInput{
		ID:                "w1",
		Status:            "done",
		Title:             "P1 Task One",
		Description:       "Updated",
		DescriptionFormat: "markdown",
		DryRun:            true,
	})
	if err != nil {
		t.Fatalf("expected update dry-run to succeed, got: %v", err)
	}
	if !updated.DryRun || workitems.updateCalled {
		t.Fatalf("expected dry-run without mutation, result=%+v called=%v", updated, workitems.updateCalled)
	}
}

func TestWorkitemCreateMissingAssigneeErrorAddsHint(t *testing.T) {
	workitems := &fakeWorkitemService{createErr: fmt.Errorf("指派人不能为空")}
	useCase := NewWorkitemUseCase(&fakeProjectService{}, workitems, nil, safety.Environment{})

	_, err := useCase.CreateWorkitem(context.Background(), CreateWorkitemInput{
		ProjectID: "p1",
		Type:      "task",
		Title:     "Task One",
	})
	if err == nil {
		t.Fatal("expected create error")
	}
	if !strings.Contains(err.Error(), "--assignee @me") || !strings.Contains(err.Error(), "yx member search --name") {
		t.Fatalf("expected actionable missing assignee hint, got: %v", err)
	}
}

func TestWorkitemAssigneeResolver(t *testing.T) {
	workitems := &fakeWorkitemService{}
	resolver := &fakeAssigneeResolver{resolved: "u-me"}
	useCase := NewWorkitemUseCaseWithAssigneeResolver(&fakeProjectService{}, workitems, resolver, nil, safety.Environment{})

	_, err := useCase.CreateWorkitem(context.Background(), CreateWorkitemInput{
		ProjectID: "p1",
		Type:      "task",
		Title:     "Task One",
		Assignee:  "@me",
	})
	if err != nil {
		t.Fatalf("create workitem: %v", err)
	}
	if resolver.input != "@me" || workitems.createInput.Assignee != "u-me" {
		t.Fatalf("expected resolved assignee, resolver=%+v input=%+v", resolver, workitems.createInput)
	}
}

func TestWorkitemDeleteHonorsSafety(t *testing.T) {
	workitems := &fakeWorkitemService{}
	useCase := NewWorkitemUseCase(&fakeProjectService{}, workitems, nil, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	result, err := useCase.DeleteWorkitem(context.Background(), DeleteWorkitemInput{ID: "w1", DryRun: true})
	if err != nil {
		t.Fatalf("expected delete dry-run to succeed, got: %v", err)
	}
	if !result.DryRun || result.Summary != "delete workitem w1" {
		t.Fatalf("unexpected dry-run result: %+v", result)
	}
	if workitems.deleteCalled {
		t.Fatal("expected dry-run not to call delete")
	}

	_, err = useCase.DeleteWorkitem(context.Background(), DeleteWorkitemInput{ID: "w1"})
	if err == nil {
		t.Fatal("expected delete without yes or dry-run to require confirmation")
	}
	if workitems.deleteCalled {
		t.Fatal("expected delete not to be called without confirmation")
	}

	result, err = useCase.DeleteWorkitem(context.Background(), DeleteWorkitemInput{ID: "w1", Yes: true})
	if err != nil {
		t.Fatalf("expected delete with yes to succeed, got: %v", err)
	}
	if result.Workitem.ID != "w1" {
		t.Fatalf("unexpected delete result: %+v", result)
	}
}

type fakeProjectService struct {
	projects     []Project
	templates    []ProjectTemplate
	created      Project
	createInput  CreateProjectInput
	createCalled bool
}

func (s *fakeProjectService) ListProjects(ctx context.Context) ([]Project, error) {
	return s.projects, nil
}

func (s *fakeProjectService) ListProjectTemplates(ctx context.Context) ([]ProjectTemplate, error) {
	return s.templates, nil
}

func (s *fakeProjectService) CreateProject(ctx context.Context, input CreateProjectInput) (Project, error) {
	s.createCalled = true
	s.createInput = input
	if s.created.ID == "" {
		return Project{ID: "p1", Name: input.Name, CustomCode: input.CustomCode}, nil
	}
	return s.created, nil
}

type fakeWorkitemService struct {
	list          []WorkitemListItem
	detail        WorkitemDetail
	created       WorkitemDetail
	createInput   CreateWorkitemInput
	updated       WorkitemDetail
	lastProjectID string
	listCalled    bool
	createCalled  bool
	createErr     error
	updateCalled  bool
	deleteCalled  bool
}

func (s *fakeWorkitemService) ListWorkitems(ctx context.Context, projectID string) ([]WorkitemListItem, error) {
	s.listCalled = true
	s.lastProjectID = projectID
	return s.list, nil
}

func (s *fakeWorkitemService) GetWorkitem(ctx context.Context, id string) (WorkitemDetail, error) {
	return s.detail, nil
}

func (s *fakeWorkitemService) CreateWorkitem(ctx context.Context, input CreateWorkitemInput) (WorkitemDetail, error) {
	s.createCalled = true
	s.createInput = input
	if s.createErr != nil {
		return WorkitemDetail{}, s.createErr
	}
	if s.created.ID == "" {
		return WorkitemDetail{ID: "w1", Title: input.Title, Type: input.Type, ProjectID: input.ProjectID}, nil
	}
	return s.created, nil
}

func (s *fakeWorkitemService) UpdateWorkitem(ctx context.Context, input UpdateWorkitemInput) (WorkitemDetail, error) {
	s.updateCalled = true
	if s.updated.ID == "" {
		return WorkitemDetail{ID: input.ID, Status: input.Status, Assignee: input.Assignee}, nil
	}
	return s.updated, nil
}

func (s *fakeWorkitemService) DeleteWorkitem(ctx context.Context, id string) (WorkitemDetail, error) {
	s.deleteCalled = true
	return WorkitemDetail{ID: id}, nil
}

type fakeAssigneeResolver struct {
	input    string
	resolved string
}

func (r *fakeAssigneeResolver) ResolveAssignee(ctx context.Context, assignee string) (string, error) {
	r.input = assignee
	return r.resolved, nil
}
