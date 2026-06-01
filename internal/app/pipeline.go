package app

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

type PipelineListItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PipelineDetail struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PipelineRun struct {
	ID         string          `json:"id"`
	PipelineID string          `json:"pipelineId"`
	Status     string          `json:"status"`
	Branch     string          `json:"branch,omitempty"`
	Tag        string          `json:"tag,omitempty"`
	CommitID   string          `json:"commitId,omitempty"`
	Trigger    string          `json:"trigger,omitempty"`
	CreatedAt  string          `json:"createdAt,omitempty"`
	StartedAt  string          `json:"startedAt,omitempty"`
	FinishedAt string          `json:"finishedAt,omitempty"`
	Stages     []PipelineStage `json:"stages,omitempty"`
	Jobs       []PipelineJob   `json:"jobs,omitempty"`
}

type PipelineStage struct {
	ID     string        `json:"id,omitempty"`
	Name   string        `json:"name"`
	Status string        `json:"status"`
	Jobs   []PipelineJob `json:"jobs,omitempty"`
}

type PipelineJob struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PipelineJobStep struct {
	StepIndex string `json:"stepIndex"`
	BuildID   string `json:"buildId"`
	Name      string `json:"name"`
	Status    string `json:"status"`
}

type PipelineCreateInput struct {
	Name    string
	Content string
	DryRun  bool
	Yes     bool
}

type PipelineRunInput struct {
	PipelineID string
	Branch     string
	DryRun     bool
	Yes        bool
}

type PipelineRunListInput struct {
	PipelineID string
	Branch     string
	Tag        string
	Commit     string
	Page       int
	PerPage    int
}

type PipelineRunGetInput struct {
	PipelineID string
	RunID      string
}

type PipelineJobRunLogInput struct {
	PipelineID string
	RunID      string
	JobID      string
	StepIndex  string
	BuildID    string
	Offset     int
	Limit      int
}

type PipelineJobRunLog struct {
	Content string `json:"content"`
	Last    int    `json:"last,omitempty"`
	More    bool   `json:"more"`
}

type PipelineLogsInput struct {
	RunID  string
	Follow bool
}

type PipelineRunResult struct {
	DryRun  bool        `json:"dryRun"`
	Summary string      `json:"summary,omitempty"`
	Run     PipelineRun `json:"run,omitempty"`
}

type PipelineMutationResult struct {
	DryRun   bool           `json:"dryRun"`
	Summary  string         `json:"summary,omitempty"`
	Pipeline PipelineDetail `json:"pipeline,omitempty"`
}

type PipelineService interface {
	ListPipelines(ctx context.Context) ([]PipelineListItem, error)
	GetPipeline(ctx context.Context, id string) (PipelineDetail, error)
	CreatePipeline(ctx context.Context, input PipelineCreateInput) (PipelineDetail, error)
	RunPipeline(ctx context.Context, input PipelineRunInput) (PipelineRun, error)
	ListPipelineRuns(ctx context.Context, input PipelineRunListInput) ([]PipelineRun, error)
	GetPipelineRun(ctx context.Context, input PipelineRunGetInput) (PipelineRun, error)
	GetPipelineJobSteps(ctx context.Context, input PipelineJobRunLogInput) ([]PipelineJobStep, error)
	GetPipelineJobRunLog(ctx context.Context, input PipelineJobRunLogInput) (PipelineJobRunLog, error)
	GetPipelineLogs(ctx context.Context, input PipelineLogsInput) ([]string, error)
}

type PipelineUseCase struct {
	service PipelineService
	safety  safety.Environment
}

func NewPipelineUseCase(service PipelineService, safetyEnv safety.Environment) *PipelineUseCase {
	return &PipelineUseCase{service: service, safety: safetyEnv}
}

func (u *PipelineUseCase) ListPipelines(ctx context.Context) ([]PipelineListItem, error) {
	return u.service.ListPipelines(ctx)
}

func (u *PipelineUseCase) GetPipeline(ctx context.Context, id string) (PipelineDetail, error) {
	if id == "" {
		return PipelineDetail{}, fmt.Errorf("pipeline id is required")
	}
	return u.service.GetPipeline(ctx, id)
}

func (u *PipelineUseCase) CreatePipeline(ctx context.Context, input PipelineCreateInput) (PipelineMutationResult, error) {
	if input.Name == "" || input.Content == "" {
		return PipelineMutationResult{}, fmt.Errorf("pipeline name and content are required")
	}
	summary := fmt.Sprintf("create pipeline %s", input.Name)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return PipelineMutationResult{}, err
	}
	if decision.DryRun {
		return PipelineMutationResult{DryRun: true, Summary: summary}, nil
	}
	pipeline, err := u.service.CreatePipeline(ctx, input)
	if err != nil {
		return PipelineMutationResult{}, err
	}
	return PipelineMutationResult{Pipeline: pipeline}, nil
}

func (u *PipelineUseCase) RunPipeline(ctx context.Context, input PipelineRunInput) (PipelineRunResult, error) {
	if input.PipelineID == "" || input.Branch == "" {
		return PipelineRunResult{}, fmt.Errorf("pipeline id and branch are required")
	}
	summary := fmt.Sprintf("run pipeline %s on %s", input.PipelineID, input.Branch)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return PipelineRunResult{}, err
	}
	if decision.DryRun {
		return PipelineRunResult{DryRun: true, Summary: summary}, nil
	}
	run, err := u.service.RunPipeline(ctx, input)
	if err != nil {
		return PipelineRunResult{}, err
	}
	return PipelineRunResult{Run: run}, nil
}

func (u *PipelineUseCase) ListPipelineRuns(ctx context.Context, input PipelineRunListInput) ([]PipelineRun, error) {
	if input.PipelineID == "" {
		return nil, fmt.Errorf("pipeline id is required")
	}
	return u.service.ListPipelineRuns(ctx, input)
}

func (u *PipelineUseCase) GetPipelineRun(ctx context.Context, input PipelineRunGetInput) (PipelineRun, error) {
	if input.PipelineID == "" || input.RunID == "" {
		return PipelineRun{}, fmt.Errorf("pipeline id and run id are required")
	}
	return u.service.GetPipelineRun(ctx, input)
}

func (u *PipelineUseCase) GetPipelineJobRunLog(ctx context.Context, input PipelineJobRunLogInput) (PipelineJobRunLog, error) {
	if input.PipelineID == "" || input.RunID == "" || input.JobID == "" {
		return PipelineJobRunLog{}, fmt.Errorf("pipeline id, run id, and job id are required")
	}
	if (input.StepIndex == "") != (input.BuildID == "") {
		return PipelineJobRunLog{}, fmt.Errorf("step index and build id must be provided together")
	}
	if input.Offset < 0 || input.Limit < 0 {
		return PipelineJobRunLog{}, fmt.Errorf("offset and limit must be non-negative")
	}
	return u.service.GetPipelineJobRunLog(ctx, input)
}

func (u *PipelineUseCase) GetPipelineJobSteps(ctx context.Context, input PipelineJobRunLogInput) ([]PipelineJobStep, error) {
	if input.PipelineID == "" || input.RunID == "" || input.JobID == "" {
		return nil, fmt.Errorf("pipeline id, run id, and job id are required")
	}
	return u.service.GetPipelineJobSteps(ctx, input)
}

func (u *PipelineUseCase) GetPipelineLogs(ctx context.Context, input PipelineLogsInput) ([]string, error) {
	if input.RunID == "" {
		return nil, fmt.Errorf("pipeline run id is required")
	}
	return u.service.GetPipelineLogs(ctx, input)
}
