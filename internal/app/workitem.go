package app

import (
	"context"
	"fmt"
	"time"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CustomCode string `json:"customCode,omitempty"`
	Scope      string `json:"scope,omitempty"`
}

type ProjectTemplate struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WorkitemListItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	ProjectID string `json:"projectId"`
}

type WorkitemDetail struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	ProjectID string `json:"projectId"`
	Assignee  string `json:"assignee,omitempty"`
}

type WorkitemListInput struct {
	ProjectID string
	Repo      string
}

type CreateProjectInput struct {
	Name        string
	CustomCode  string
	Scope       string
	TemplateID  string
	Description string
	DryRun      bool
	Yes         bool
}

type CreateWorkitemInput struct {
	ProjectID string
	Type      string
	Title     string
	DryRun    bool
	Yes       bool
}

type UpdateWorkitemInput struct {
	ID       string
	Status   string
	Assignee string
	DryRun   bool
	Yes      bool
}

type WorkitemMutationResult struct {
	DryRun   bool           `json:"dryRun"`
	Summary  string         `json:"summary,omitempty"`
	Workitem WorkitemDetail `json:"workitem,omitempty"`
}

type ProjectMutationResult struct {
	DryRun  bool    `json:"dryRun"`
	Summary string  `json:"summary,omitempty"`
	Project Project `json:"project,omitempty"`
}

type ProjectService interface {
	ListProjects(ctx context.Context) ([]Project, error)
	ListProjectTemplates(ctx context.Context) ([]ProjectTemplate, error)
	CreateProject(ctx context.Context, input CreateProjectInput) (Project, error)
}

type WorkitemService interface {
	ListWorkitems(ctx context.Context, projectID string) ([]WorkitemListItem, error)
	GetWorkitem(ctx context.Context, id string) (WorkitemDetail, error)
	CreateWorkitem(ctx context.Context, input CreateWorkitemInput) (WorkitemDetail, error)
	UpdateWorkitem(ctx context.Context, input UpdateWorkitemInput) (WorkitemDetail, error)
}

type WorkitemUseCase struct {
	projects       ProjectService
	workitems      WorkitemService
	repoProjectMap map[string]string
	safety         safety.Environment
}

type ErrMissingRepoProjectMapping struct {
	Repo string
}

func (e ErrMissingRepoProjectMapping) Error() string {
	return fmt.Sprintf("repo %q is not mapped to a project; run yx config set repo.%s.project <project-id>", e.Repo, e.Repo)
}

func NewWorkitemUseCase(projects ProjectService, workitems WorkitemService, repoProjectMap map[string]string, safetyEnv safety.Environment) *WorkitemUseCase {
	if repoProjectMap == nil {
		repoProjectMap = map[string]string{}
	}
	return &WorkitemUseCase{
		projects:       projects,
		workitems:      workitems,
		repoProjectMap: repoProjectMap,
		safety:         safetyEnv,
	}
}

func (u *WorkitemUseCase) ListProjects(ctx context.Context) ([]Project, error) {
	return u.projects.ListProjects(ctx)
}

func (u *WorkitemUseCase) CreateProject(ctx context.Context, input CreateProjectInput) (ProjectMutationResult, error) {
	if input.Name == "" {
		return ProjectMutationResult{}, fmt.Errorf("name is required")
	}
	if input.Scope == "" {
		input.Scope = "public"
	}
	if input.CustomCode == "" {
		input.CustomCode = generateProjectCustomCode(time.Now())
	}
	if input.TemplateID == "" {
		templates, err := u.projects.ListProjectTemplates(ctx)
		if err != nil {
			return ProjectMutationResult{}, err
		}
		if len(templates) == 0 {
			return ProjectMutationResult{}, fmt.Errorf("no project templates available")
		}
		input.TemplateID = templates[0].ID
	}

	summary := fmt.Sprintf("create project %q with template %s", input.Name, input.TemplateID)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return ProjectMutationResult{}, err
	}
	if decision.DryRun {
		return ProjectMutationResult{DryRun: true, Summary: summary}, nil
	}
	project, err := u.projects.CreateProject(ctx, input)
	if err != nil {
		return ProjectMutationResult{}, err
	}
	if project.Name == "" {
		project.Name = input.Name
	}
	if project.CustomCode == "" {
		project.CustomCode = input.CustomCode
	}
	if project.Scope == "" {
		project.Scope = input.Scope
	}
	return ProjectMutationResult{Project: project}, nil
}

func (u *WorkitemUseCase) ListWorkitems(ctx context.Context, input WorkitemListInput) ([]WorkitemListItem, error) {
	projectID := input.ProjectID
	if projectID == "" && input.Repo != "" {
		mapped, ok := u.repoProjectMap[input.Repo]
		if !ok {
			return nil, ErrMissingRepoProjectMapping{Repo: input.Repo}
		}
		projectID = mapped
	}
	if projectID == "" {
		return nil, fmt.Errorf("project is required")
	}
	return u.workitems.ListWorkitems(ctx, projectID)
}

func (u *WorkitemUseCase) GetWorkitem(ctx context.Context, id string) (WorkitemDetail, error) {
	if id == "" {
		return WorkitemDetail{}, fmt.Errorf("workitem id is required")
	}
	return u.workitems.GetWorkitem(ctx, id)
}

func (u *WorkitemUseCase) CreateWorkitem(ctx context.Context, input CreateWorkitemInput) (WorkitemMutationResult, error) {
	if input.ProjectID == "" || input.Type == "" || input.Title == "" {
		return WorkitemMutationResult{}, fmt.Errorf("project, type, and title are required")
	}
	summary := fmt.Sprintf("create %s workitem %q in %s", input.Type, input.Title, input.ProjectID)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	if decision.DryRun {
		return WorkitemMutationResult{DryRun: true, Summary: summary}, nil
	}
	detail, err := u.workitems.CreateWorkitem(ctx, input)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	return WorkitemMutationResult{Workitem: detail}, nil
}

func generateProjectCustomCode(now time.Time) string {
	n := now.UnixNano()
	code := make([]byte, 5)
	for i := range code {
		code[i] = byte('A' + n%26)
		n /= 26
	}
	return string(code)
}

func (u *WorkitemUseCase) UpdateWorkitem(ctx context.Context, input UpdateWorkitemInput) (WorkitemMutationResult, error) {
	if input.ID == "" {
		return WorkitemMutationResult{}, fmt.Errorf("workitem id is required")
	}
	if input.Status == "" && input.Assignee == "" {
		return WorkitemMutationResult{}, fmt.Errorf("status or assignee is required")
	}
	summary := fmt.Sprintf("update workitem %s", input.ID)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	if decision.DryRun {
		return WorkitemMutationResult{DryRun: true, Summary: summary}, nil
	}
	detail, err := u.workitems.UpdateWorkitem(ctx, input)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	return WorkitemMutationResult{Workitem: detail}, nil
}
