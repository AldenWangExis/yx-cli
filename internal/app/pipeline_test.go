package app

import (
	"context"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

func TestPipelineUseCaseListViewAndLogs(t *testing.T) {
	service := &fakePipelineService{
		list:   []PipelineListItem{{ID: "pipe1", Name: "Build", Status: "enabled"}},
		detail: PipelineDetail{ID: "pipe1", Name: "Build", Status: "enabled"},
		logs:   []string{"line 1", "line 2"},
	}
	useCase := NewPipelineUseCase(service, safety.Environment{})

	listed, err := useCase.ListPipelines(context.Background())
	if err != nil {
		t.Fatalf("expected list to succeed, got: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != "pipe1" {
		t.Fatalf("unexpected pipelines: %+v", listed)
	}

	detail, err := useCase.GetPipeline(context.Background(), "pipe1")
	if err != nil {
		t.Fatalf("expected detail to succeed, got: %v", err)
	}
	if detail.Name != "Build" {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	logs, err := useCase.GetPipelineLogs(context.Background(), PipelineLogsInput{RunID: "run1", Follow: true})
	if err != nil {
		t.Fatalf("expected logs to succeed, got: %v", err)
	}
	if len(logs) != 2 || !service.follow {
		t.Fatalf("unexpected logs=%+v follow=%v", logs, service.follow)
	}
}

func TestPipelineRunDryRunDoesNotMutate(t *testing.T) {
	service := &fakePipelineService{}
	useCase := NewPipelineUseCase(service, safety.Environment{ConfirmWrites: true})

	result, err := useCase.RunPipeline(context.Background(), PipelineRunInput{
		PipelineID: "pipe1",
		Branch:     "main",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("expected dry run to succeed, got: %v", err)
	}
	if !result.DryRun || service.runCalled {
		t.Fatalf("expected dry-run without mutation, result=%+v called=%v", result, service.runCalled)
	}
}

func TestPipelineRunRequiresConfirmation(t *testing.T) {
	service := &fakePipelineService{}
	useCase := NewPipelineUseCase(service, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	_, err := useCase.RunPipeline(context.Background(), PipelineRunInput{PipelineID: "pipe1", Branch: "main"})
	if err == nil {
		t.Fatal("expected run to require confirmation")
	}
	if service.runCalled {
		t.Fatal("expected run not to be called without confirmation")
	}
}

func TestPipelineRunWithYesMutatesOnce(t *testing.T) {
	service := &fakePipelineService{run: PipelineRun{ID: "run1", PipelineID: "pipe1", Status: "running"}}
	useCase := NewPipelineUseCase(service, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	result, err := useCase.RunPipeline(context.Background(), PipelineRunInput{PipelineID: "pipe1", Branch: "main", Yes: true})
	if err != nil {
		t.Fatalf("expected run with yes to succeed, got: %v", err)
	}
	if result.Run.ID != "run1" {
		t.Fatalf("unexpected run result: %+v", result)
	}
	if service.runCalls != 1 {
		t.Fatalf("expected one run call, got %d", service.runCalls)
	}
}

type fakePipelineService struct {
	list      []PipelineListItem
	detail    PipelineDetail
	run       PipelineRun
	logs      []string
	follow    bool
	runCalled bool
	runCalls  int
}

func (s *fakePipelineService) ListPipelines(ctx context.Context) ([]PipelineListItem, error) {
	return s.list, nil
}

func (s *fakePipelineService) GetPipeline(ctx context.Context, id string) (PipelineDetail, error) {
	return s.detail, nil
}

func (s *fakePipelineService) RunPipeline(ctx context.Context, input PipelineRunInput) (PipelineRun, error) {
	s.runCalled = true
	s.runCalls++
	if s.run.ID == "" {
		return PipelineRun{ID: "run1", PipelineID: input.PipelineID, Status: "running"}, nil
	}
	return s.run, nil
}

func (s *fakePipelineService) GetPipelineLogs(ctx context.Context, input PipelineLogsInput) ([]string, error) {
	s.follow = input.Follow
	return s.logs, nil
}
