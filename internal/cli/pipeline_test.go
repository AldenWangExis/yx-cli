package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

func TestPipelineCommands(t *testing.T) {
	pipelines := &fakePipelineUseCase{
		list:   []app.PipelineListItem{{ID: "pipe1", Name: "Build", Status: "enabled"}},
		detail: app.PipelineDetail{ID: "pipe1", Name: "Build", Status: "enabled"},
		create: app.PipelineMutationResult{DryRun: true, Summary: "create pipeline yx-cli-ci"},
		run:    app.PipelineRunResult{DryRun: true, Summary: "run pipeline pipe1 on main"},
		runs:   []app.PipelineRun{{ID: "run1", PipelineID: "pipe1", Status: "SUCCESS", Branch: "main", CommitID: "abc123"}},
		runDetail: app.PipelineRun{
			ID:         "run1",
			PipelineID: "pipe1",
			Status:     "SUCCESS",
			Jobs:       []app.PipelineJob{{ID: "job1", Name: "Run tests", Status: "SUCCESS"}},
		},
		jobSteps: []app.PipelineJobStep{{StepIndex: "1", BuildID: "99", Name: "Run go test", Status: "FAIL"}},
		jobLogs:  app.PipelineJobRunLog{Content: "line 1\nline 2", Last: 2, More: false},
		logs:     []string{"line 1", "line 2"},
	}
	opts := Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml"), PipelineUseCase: pipelines}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "--json", "pipeline", "list")
	if err != nil {
		t.Fatalf("expected pipeline list, got: %v stderr=%s", err, stderr)
	}
	var listed []app.PipelineListItem
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatalf("expected JSON list, got: %v output=%s", err, stdout)
	}
	if len(listed) != 1 || listed[0].ID != "pipe1" {
		t.Fatalf("unexpected list: %+v", listed)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "pipeline", "view", "pipe1")
	if err != nil {
		t.Fatalf("expected pipeline view, got: %v stderr=%s", err, stderr)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "pipeline", "run", "pipe1", "--branch", "main", "--dry-run")
	if err != nil {
		t.Fatalf("expected pipeline run dry-run, got: %v stderr=%s", err, stderr)
	}
	if pipelines.runInput.PipelineID != "pipe1" || pipelines.runInput.Branch != "main" || !pipelines.runInput.DryRun {
		t.Fatalf("unexpected run input: %+v", pipelines.runInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "pipeline", "run", "list", "pipe1", "--branch", "main", "--tag", "v1.0.0-alpha", "--commit", "abc123")
	if err != nil {
		t.Fatalf("expected pipeline run list, got: %v stderr=%s", err, stderr)
	}
	var runs []app.PipelineRun
	if err := json.Unmarshal([]byte(stdout), &runs); err != nil {
		t.Fatalf("expected JSON run list, got: %v output=%s", err, stdout)
	}
	if len(runs) != 1 || pipelines.runListInput.Tag != "v1.0.0-alpha" {
		t.Fatalf("unexpected run list=%+v input=%+v", runs, pipelines.runListInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "pipeline", "run", "view", "pipe1", "run1")
	if err != nil {
		t.Fatalf("expected pipeline run view, got: %v stderr=%s", err, stderr)
	}
	var runDetail app.PipelineRun
	if err := json.Unmarshal([]byte(stdout), &runDetail); err != nil {
		t.Fatalf("expected JSON run detail, got: %v output=%s", err, stdout)
	}
	if len(runDetail.Jobs) != 1 || pipelines.runGetInput.RunID != "run1" {
		t.Fatalf("unexpected run detail=%+v input=%+v", runDetail, pipelines.runGetInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "--json", "pipeline", "run", "steps", "pipe1", "run1", "--job", "job1")
	if err != nil {
		t.Fatalf("expected pipeline run steps, got: %v stderr=%s", err, stderr)
	}
	var steps []app.PipelineJobStep
	if err := json.Unmarshal([]byte(stdout), &steps); err != nil {
		t.Fatalf("expected JSON steps, got: %v output=%s", err, stdout)
	}
	if len(steps) != 1 || pipelines.jobStepsInput.JobID != "job1" {
		t.Fatalf("unexpected job steps=%+v input=%+v", steps, pipelines.jobStepsInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "pipeline", "run", "logs", "pipe1", "run1", "--job", "job1", "--step-index", "1", "--build-id", "99", "--offset", "2", "--limit", "200")
	if err != nil {
		t.Fatalf("expected pipeline run logs, got: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "line 1") || pipelines.jobLogInput.JobID != "job1" || pipelines.jobLogInput.StepIndex != "1" || pipelines.jobLogInput.BuildID != "99" || pipelines.jobLogInput.Offset != 2 || pipelines.jobLogInput.Limit != 200 {
		t.Fatalf("unexpected job logs output=%s input=%+v", stdout, pipelines.jobLogInput)
	}

	contentPath := filepath.Join(t.TempDir(), "flow.yml")
	if err := os.WriteFile(contentPath, []byte("stages: []\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "pipeline", "create", "--name", "yx-cli-ci", "--file", contentPath, "--dry-run")
	if err != nil {
		t.Fatalf("expected pipeline create dry-run, got: %v stderr=%s", err, stderr)
	}
	if pipelines.createInput.Name != "yx-cli-ci" || pipelines.createInput.Content != "stages: []\n" || !pipelines.createInput.DryRun {
		t.Fatalf("unexpected create input: %+v", pipelines.createInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "pipeline", "logs", "run1", "--follow")
	if err != nil {
		t.Fatalf("expected pipeline logs, got: %v stderr=%s", err, stderr)
	}
	if !pipelines.logsInput.Follow {
		t.Fatal("expected follow flag")
	}
	if !strings.Contains(stdout, "line 1") || !strings.Contains(stdout, "line 2") {
		t.Fatalf("unexpected logs output: %s", stdout)
	}
}

type fakePipelineUseCase struct {
	list          []app.PipelineListItem
	detail        app.PipelineDetail
	create        app.PipelineMutationResult
	run           app.PipelineRunResult
	runs          []app.PipelineRun
	runDetail     app.PipelineRun
	jobSteps      []app.PipelineJobStep
	jobLogs       app.PipelineJobRunLog
	logs          []string
	createInput   app.PipelineCreateInput
	runInput      app.PipelineRunInput
	runListInput  app.PipelineRunListInput
	runGetInput   app.PipelineRunGetInput
	jobStepsInput app.PipelineJobRunLogInput
	jobLogInput   app.PipelineJobRunLogInput
	logsInput     app.PipelineLogsInput
}

func (u *fakePipelineUseCase) ListPipelines(ctx context.Context) ([]app.PipelineListItem, error) {
	return u.list, nil
}

func (u *fakePipelineUseCase) GetPipeline(ctx context.Context, id string) (app.PipelineDetail, error) {
	return u.detail, nil
}

func (u *fakePipelineUseCase) CreatePipeline(ctx context.Context, input app.PipelineCreateInput) (app.PipelineMutationResult, error) {
	u.createInput = input
	return u.create, nil
}

func (u *fakePipelineUseCase) RunPipeline(ctx context.Context, input app.PipelineRunInput) (app.PipelineRunResult, error) {
	u.runInput = input
	return u.run, nil
}

func (u *fakePipelineUseCase) ListPipelineRuns(ctx context.Context, input app.PipelineRunListInput) ([]app.PipelineRun, error) {
	u.runListInput = input
	return u.runs, nil
}

func (u *fakePipelineUseCase) GetPipelineRun(ctx context.Context, input app.PipelineRunGetInput) (app.PipelineRun, error) {
	u.runGetInput = input
	return u.runDetail, nil
}

func (u *fakePipelineUseCase) GetPipelineJobRunLog(ctx context.Context, input app.PipelineJobRunLogInput) (app.PipelineJobRunLog, error) {
	u.jobLogInput = input
	return u.jobLogs, nil
}

func (u *fakePipelineUseCase) GetPipelineJobSteps(ctx context.Context, input app.PipelineJobRunLogInput) ([]app.PipelineJobStep, error) {
	u.jobStepsInput = input
	return u.jobSteps, nil
}

func (u *fakePipelineUseCase) GetPipelineLogs(ctx context.Context, input app.PipelineLogsInput) ([]string, error) {
	u.logsInput = input
	return u.logs, nil
}
