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
	ID         string `json:"id"`
	PipelineID string `json:"pipelineId"`
	Status     string `json:"status"`
}

type PipelineRunInput struct {
	PipelineID string
	Branch     string
	DryRun     bool
	Yes        bool
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

type PipelineService interface {
	ListPipelines(ctx context.Context) ([]PipelineListItem, error)
	GetPipeline(ctx context.Context, id string) (PipelineDetail, error)
	RunPipeline(ctx context.Context, input PipelineRunInput) (PipelineRun, error)
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

func (u *PipelineUseCase) GetPipelineLogs(ctx context.Context, input PipelineLogsInput) ([]string, error) {
	if input.RunID == "" {
		return nil, fmt.Errorf("pipeline run id is required")
	}
	return u.service.GetPipelineLogs(ctx, input)
}
