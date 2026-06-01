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
		runs:   []PipelineRun{{ID: "run1", PipelineID: "pipe1", Status: "SUCCESS", Branch: "main", CommitID: "abc123"}},
		runDetail: PipelineRun{
			ID:         "run1",
			PipelineID: "pipe1",
			Status:     "SUCCESS",
			Jobs:       []PipelineJob{{ID: "job1", Name: "Run tests", Status: "SUCCESS"}},
		},
		jobSteps: []PipelineJobStep{{StepIndex: "1", BuildID: "99", Name: "Run go test", Status: "FAIL"}},
		jobLogs:  PipelineJobRunLog{Content: "line 1\nline 2", Last: 2, More: false},
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

	runs, err := useCase.ListPipelineRuns(context.Background(), PipelineRunListInput{
		PipelineID: "pipe1",
		Branch:     "main",
		Tag:        "v1.0.0-alpha",
		Commit:     "abc123",
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("expected pipeline runs to list, got: %v", err)
	}
	if len(runs) != 1 || runs[0].ID != "run1" || service.runListInput.Tag != "v1.0.0-alpha" {
		t.Fatalf("unexpected runs=%+v input=%+v", runs, service.runListInput)
	}

	run, err := useCase.GetPipelineRun(context.Background(), PipelineRunGetInput{PipelineID: "pipe1", RunID: "run1"})
	if err != nil {
		t.Fatalf("expected run detail, got: %v", err)
	}
	if len(run.Jobs) != 1 || run.Jobs[0].ID != "job1" {
		t.Fatalf("unexpected run detail: %+v", run)
	}

	jobLogs, err := useCase.GetPipelineJobRunLog(context.Background(), PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1"})
	if err != nil {
		t.Fatalf("expected job logs, got: %v", err)
	}
	if jobLogs.Content != "line 1\nline 2" || service.jobLogInput.JobID != "job1" {
		t.Fatalf("unexpected job logs=%+v input=%+v", jobLogs, service.jobLogInput)
	}

	steps, err := useCase.GetPipelineJobSteps(context.Background(), PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1"})
	if err != nil {
		t.Fatalf("expected job steps, got: %v", err)
	}
	if len(steps) != 1 || steps[0].BuildID != "99" || !service.jobStepsCalled {
		t.Fatalf("unexpected job steps=%+v input=%+v", steps, service.jobStepsInput)
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

func TestPipelineCreateDryRunDoesNotMutate(t *testing.T) {
	service := &fakePipelineService{}
	useCase := NewPipelineUseCase(service, safety.Environment{ConfirmWrites: true})

	result, err := useCase.CreatePipeline(context.Background(), PipelineCreateInput{
		Name:    "yx-cli-ci",
		Content: "stages: []\n",
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("expected dry run to succeed, got: %v", err)
	}
	if !result.DryRun || service.createCalled {
		t.Fatalf("expected dry-run without mutation, result=%+v called=%v", result, service.createCalled)
	}
}

func TestPipelineCreateWithYesMutatesOnce(t *testing.T) {
	service := &fakePipelineService{created: PipelineDetail{ID: "pipe2", Name: "yx-cli-ci", Status: "enabled"}}
	useCase := NewPipelineUseCase(service, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	result, err := useCase.CreatePipeline(context.Background(), PipelineCreateInput{
		Name:    "yx-cli-ci",
		Content: "stages: []\n",
		Yes:     true,
	})
	if err != nil {
		t.Fatalf("expected create with yes to succeed, got: %v", err)
	}
	if result.Pipeline.ID != "pipe2" {
		t.Fatalf("unexpected create result: %+v", result)
	}
	if service.createCalls != 1 || service.createInput.Content != "stages: []\n" {
		t.Fatalf("unexpected create call: calls=%d input=%+v", service.createCalls, service.createInput)
	}
}

func TestPipelineCreateRequiresNameAndContent(t *testing.T) {
	service := &fakePipelineService{}
	useCase := NewPipelineUseCase(service, safety.Environment{})

	_, err := useCase.CreatePipeline(context.Background(), PipelineCreateInput{Name: "yx-cli-ci"})
	if err == nil {
		t.Fatal("expected create without content to fail")
	}
	if service.createCalled {
		t.Fatal("expected create not to be called without content")
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

func TestPipelineRunQueriesRequireIdentifiers(t *testing.T) {
	service := &fakePipelineService{}
	useCase := NewPipelineUseCase(service, safety.Environment{})

	if _, err := useCase.ListPipelineRuns(context.Background(), PipelineRunListInput{}); err == nil {
		t.Fatal("expected list runs without pipeline id to fail")
	}
	if _, err := useCase.GetPipelineRun(context.Background(), PipelineRunGetInput{PipelineID: "pipe1"}); err == nil {
		t.Fatal("expected get run without run id to fail")
	}
	if _, err := useCase.GetPipelineJobRunLog(context.Background(), PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1"}); err == nil {
		t.Fatal("expected job logs without job id to fail")
	}
	if _, err := useCase.GetPipelineJobRunLog(context.Background(), PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1", StepIndex: "1"}); err == nil {
		t.Fatal("expected job logs without build id to fail")
	}
	if _, err := useCase.GetPipelineJobRunLog(context.Background(), PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1", BuildID: "99"}); err == nil {
		t.Fatal("expected job logs without step index to fail")
	}
	if _, err := useCase.GetPipelineJobRunLog(context.Background(), PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1", Offset: -1}); err == nil {
		t.Fatal("expected negative offset to fail")
	}
	if _, err := useCase.GetPipelineJobSteps(context.Background(), PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1"}); err == nil {
		t.Fatal("expected job steps without job id to fail")
	}
	if service.runListCalled || service.runDetailCalled || service.jobLogCalled || service.jobStepsCalled {
		t.Fatal("expected service not to be called for invalid run queries")
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
	list            []PipelineListItem
	detail          PipelineDetail
	created         PipelineDetail
	run             PipelineRun
	runs            []PipelineRun
	runDetail       PipelineRun
	logs            []string
	jobLogs         PipelineJobRunLog
	jobSteps        []PipelineJobStep
	follow          bool
	createInput     PipelineCreateInput
	runListInput    PipelineRunListInput
	runGetInput     PipelineRunGetInput
	jobLogInput     PipelineJobRunLogInput
	jobStepsInput   PipelineJobRunLogInput
	createCalled    bool
	createCalls     int
	runCalled       bool
	runCalls        int
	runListCalled   bool
	runDetailCalled bool
	jobLogCalled    bool
	jobStepsCalled  bool
}

func (s *fakePipelineService) ListPipelines(ctx context.Context) ([]PipelineListItem, error) {
	return s.list, nil
}

func (s *fakePipelineService) GetPipeline(ctx context.Context, id string) (PipelineDetail, error) {
	return s.detail, nil
}

func (s *fakePipelineService) CreatePipeline(ctx context.Context, input PipelineCreateInput) (PipelineDetail, error) {
	s.createCalled = true
	s.createCalls++
	s.createInput = input
	if s.created.ID == "" {
		return PipelineDetail{ID: "pipe2", Name: input.Name, Status: "enabled"}, nil
	}
	return s.created, nil
}

func (s *fakePipelineService) RunPipeline(ctx context.Context, input PipelineRunInput) (PipelineRun, error) {
	s.runCalled = true
	s.runCalls++
	if s.run.ID == "" {
		return PipelineRun{ID: "run1", PipelineID: input.PipelineID, Status: "running"}, nil
	}
	return s.run, nil
}

func (s *fakePipelineService) ListPipelineRuns(ctx context.Context, input PipelineRunListInput) ([]PipelineRun, error) {
	s.runListCalled = true
	s.runListInput = input
	return s.runs, nil
}

func (s *fakePipelineService) GetPipelineRun(ctx context.Context, input PipelineRunGetInput) (PipelineRun, error) {
	s.runDetailCalled = true
	s.runGetInput = input
	return s.runDetail, nil
}

func (s *fakePipelineService) GetPipelineJobRunLog(ctx context.Context, input PipelineJobRunLogInput) (PipelineJobRunLog, error) {
	s.jobLogCalled = true
	s.jobLogInput = input
	return s.jobLogs, nil
}

func (s *fakePipelineService) GetPipelineJobSteps(ctx context.Context, input PipelineJobRunLogInput) ([]PipelineJobStep, error) {
	s.jobStepsCalled = true
	s.jobStepsInput = input
	return s.jobSteps, nil
}

func (s *fakePipelineService) GetPipelineLogs(ctx context.Context, input PipelineLogsInput) ([]string, error) {
	s.follow = input.Follow
	return s.logs, nil
}
