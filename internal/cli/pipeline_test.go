package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

func TestPipelineCommands(t *testing.T) {
	pipelines := &fakePipelineUseCase{
		list:   []app.PipelineListItem{{ID: "pipe1", Name: "Build", Status: "enabled"}},
		detail: app.PipelineDetail{ID: "pipe1", Name: "Build", Status: "enabled"},
		run:    app.PipelineRunResult{DryRun: true, Summary: "run pipeline pipe1 on main"},
		logs:   []string{"line 1", "line 2"},
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
	list      []app.PipelineListItem
	detail    app.PipelineDetail
	run       app.PipelineRunResult
	logs      []string
	runInput  app.PipelineRunInput
	logsInput app.PipelineLogsInput
}

func (u *fakePipelineUseCase) ListPipelines(ctx context.Context) ([]app.PipelineListItem, error) {
	return u.list, nil
}

func (u *fakePipelineUseCase) GetPipeline(ctx context.Context, id string) (app.PipelineDetail, error) {
	return u.detail, nil
}

func (u *fakePipelineUseCase) RunPipeline(ctx context.Context, input app.PipelineRunInput) (app.PipelineRunResult, error) {
	u.runInput = input
	return u.run, nil
}

func (u *fakePipelineUseCase) GetPipelineLogs(ctx context.Context, input app.PipelineLogsInput) ([]string, error) {
	u.logsInput = input
	return u.logs, nil
}
