package app

import (
	"context"
	"fmt"
	"strings"
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
	ProjectID         string
	Type              string
	Title             string
	Description       string
	DescriptionFormat string
	Assignee          string
	DryRun            bool
	Yes               bool
}

type UpdateWorkitemInput struct {
	ID                string
	Status            string
	Assignee          string
	Title             string
	Description       string
	DescriptionFormat string
	DryRun            bool
	Yes               bool
}

type DeleteWorkitemInput struct {
	ID     string
	DryRun bool
	Yes    bool
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
	DeleteWorkitem(ctx context.Context, id string) (WorkitemDetail, error)
}

type AssigneeResolver interface {
	ResolveAssignee(ctx context.Context, assignee string) (string, error)
}

type WorkitemUseCase struct {
	projects         ProjectService
	workitems        WorkitemService
	assigneeResolver AssigneeResolver
	repoProjectMap   map[string]string
	safety           safety.Environment
}

type ErrMissingRepoProjectMapping struct {
	Repo string
}

func (e ErrMissingRepoProjectMapping) Error() string {
	return fmt.Sprintf("repo %q is not mapped to a project; run yx config set profiles.<profile>.repoProjectMap.%s <project-id>", e.Repo, e.Repo)
}

func NewWorkitemUseCase(projects ProjectService, workitems WorkitemService, repoProjectMap map[string]string, safetyEnv safety.Environment) *WorkitemUseCase {
	return NewWorkitemUseCaseWithAssigneeResolver(projects, workitems, nil, repoProjectMap, safetyEnv)
}

func NewWorkitemUseCaseWithAssigneeResolver(projects ProjectService, workitems WorkitemService, resolver AssigneeResolver, repoProjectMap map[string]string, safetyEnv safety.Environment) *WorkitemUseCase {
	if repoProjectMap == nil {
		repoProjectMap = map[string]string{}
	}
	return &WorkitemUseCase{
		projects:         projects,
		workitems:        workitems,
		assigneeResolver: resolver,
		repoProjectMap:   repoProjectMap,
		safety:           safetyEnv,
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
	if input.DescriptionFormat != "" {
		switch strings.ToLower(input.DescriptionFormat) {
		case "markdown", "richtext":
		default:
			return WorkitemMutationResult{}, fmt.Errorf("description format must be markdown or richtext")
		}
	}
	summary := fmt.Sprintf("create %s workitem %q in %s", input.Type, input.Title, input.ProjectID)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	if decision.DryRun {
		return WorkitemMutationResult{DryRun: true, Summary: summary}, nil
	}
	input, err = u.resolveCreateAssignee(ctx, input)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	detail, err := u.workitems.CreateWorkitem(ctx, input)
	if err != nil {
		return WorkitemMutationResult{}, withMissingAssigneeHint(err)
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
	if input.Status == "" && input.Assignee == "" && input.Title == "" && input.Description == "" {
		return WorkitemMutationResult{}, fmt.Errorf("status, assignee, title, or description is required")
	}
	if input.DescriptionFormat != "" {
		switch strings.ToLower(input.DescriptionFormat) {
		case "markdown", "richtext":
		default:
			return WorkitemMutationResult{}, fmt.Errorf("description format must be markdown or richtext")
		}
	}
	summary := fmt.Sprintf("update workitem %s", input.ID)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	if decision.DryRun {
		return WorkitemMutationResult{DryRun: true, Summary: summary}, nil
	}
	input, err = u.resolveUpdateAssignee(ctx, input)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	detail, err := u.workitems.UpdateWorkitem(ctx, input)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	return WorkitemMutationResult{Workitem: detail}, nil
}

func (u *WorkitemUseCase) resolveCreateAssignee(ctx context.Context, input CreateWorkitemInput) (CreateWorkitemInput, error) {
	if input.Assignee == "" || u.assigneeResolver == nil {
		return input, nil
	}
	assignee, err := u.assigneeResolver.ResolveAssignee(ctx, input.Assignee)
	if err != nil {
		return CreateWorkitemInput{}, err
	}
	input.Assignee = assignee
	return input, nil
}

func (u *WorkitemUseCase) resolveUpdateAssignee(ctx context.Context, input UpdateWorkitemInput) (UpdateWorkitemInput, error) {
	if input.Assignee == "" || u.assigneeResolver == nil {
		return input, nil
	}
	assignee, err := u.assigneeResolver.ResolveAssignee(ctx, input.Assignee)
	if err != nil {
		return UpdateWorkitemInput{}, err
	}
	input.Assignee = assignee
	return input, nil
}

func withMissingAssigneeHint(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	lower := strings.ToLower(message)
	if !strings.Contains(message, "指派人不能为空") &&
		!strings.Contains(message, "负责人不能为空") &&
		!strings.Contains(lower, "assignee") &&
		!strings.Contains(lower, "assignedto") {
		return err
	}
	return fmt.Errorf("%w\n\nThis project requires an assignee. Try assigning yourself with --assignee @me, or find a teammate with: yx member search --name <name>", err)
}

func (u *WorkitemUseCase) DeleteWorkitem(ctx context.Context, input DeleteWorkitemInput) (WorkitemMutationResult, error) {
	if input.ID == "" {
		return WorkitemMutationResult{}, fmt.Errorf("workitem id is required")
	}
	summary := fmt.Sprintf("delete workitem %s", input.ID)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	if decision.DryRun {
		return WorkitemMutationResult{DryRun: true, Summary: summary}, nil
	}
	detail, err := u.workitems.DeleteWorkitem(ctx, input.ID)
	if err != nil {
		return WorkitemMutationResult{}, err
	}
	if detail.ID == "" {
		detail.ID = input.ID
	}
	return WorkitemMutationResult{Workitem: detail}, nil
}
