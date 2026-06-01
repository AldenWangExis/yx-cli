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
	list        []app.PipelineListItem
	detail      app.PipelineDetail
	create      app.PipelineMutationResult
	run         app.PipelineRunResult
	logs        []string
	createInput app.PipelineCreateInput
	runInput    app.PipelineRunInput
	logsInput   app.PipelineLogsInput
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

func (u *fakePipelineUseCase) GetPipelineLogs(ctx context.Context, input app.PipelineLogsInput) ([]string, error) {
	u.logsInput = input
	return u.logs, nil
}
