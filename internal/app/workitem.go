package app

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

type Project struct {
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

type ProjectService interface {
	ListProjects(ctx context.Context) ([]Project, error)
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
