package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

func TestMRListAndPRAliasUseSameUseCase(t *testing.T) {
	mrs := &fakeMergeRequestUseCase{
		list: []app.MergeRequestListItem{{ID: "1", Title: "Add feature", State: "opened"}},
	}
	opts := Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml"), MergeRequestUseCase: mrs}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "--json", "mr", "list", "--repo", "repo-1")
	if err != nil {
		t.Fatalf("expected mr list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var listed []app.MergeRequestListItem
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatalf("expected mr list JSON, got error: %v output=%s", err, stdout)
	}
	if len(listed) != 1 || listed[0].Title != "Add feature" {
		t.Fatalf("unexpected merge requests: %+v", listed)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "pr", "list", "--repo", "repo-1")
	if err != nil {
		t.Fatalf("expected pr list to succeed, got error: %v stderr=%s", err, stderr)
	}
	if mrs.listCalls != 2 {
		t.Fatalf("expected mr and pr list to use same use case, got calls=%d", mrs.listCalls)
	}
}

func TestMRViewCreateMergeContracts(t *testing.T) {
	mrs := &fakeMergeRequestUseCase{
		detail: app.MergeRequestDetail{ID: "1", Title: "Add feature", State: "opened", SourceBranch: "feat", TargetBranch: "main"},
		created: app.MergeRequestMutationResult{
			DryRun:  true,
			Summary: "create merge request",
		},
		merged: app.MergeRequestMutationResult{
			DryRun:  true,
			Summary: "merge merge request",
		},
	}
	opts := Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml"), MergeRequestUseCase: mrs}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "--json", "mr", "view", "1", "--repo", "repo-1")
	if err != nil {
		t.Fatalf("expected mr view to succeed, got error: %v stderr=%s", err, stderr)
	}
	var detail app.MergeRequestDetail
	if err := json.Unmarshal([]byte(stdout), &detail); err != nil {
		t.Fatalf("expected mr detail JSON, got error: %v output=%s", err, stdout)
	}
	if detail.ID != "1" {
		t.Fatalf("expected detail id 1, got %q", detail.ID)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts),
		"mr", "create", "--repo", "repo-1", "--source", "feat", "--target", "main", "--title", "Add feature", "--dry-run")
	if err != nil {
		t.Fatalf("expected mr create dry-run to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !mrs.createInput.DryRun || mrs.createInput.SourceBranch != "feat" || mrs.createInput.TargetBranch != "main" {
		t.Fatalf("unexpected create input: %+v", mrs.createInput)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts),
		"mr", "merge", "1", "--repo", "repo-1", "--dry-run", "--yes")
	if err != nil {
		t.Fatalf("expected mr merge dry-run to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !mrs.mergeInput.DryRun || !mrs.mergeInput.Yes || mrs.mergeInput.ID != "1" {
		t.Fatalf("unexpected merge input: %+v", mrs.mergeInput)
	}
}

func TestMRListRequiresRepo(t *testing.T) {
	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		MergeRequestUseCase: &fakeMergeRequestUseCase{},
	}), "mr", "list")
	if err == nil {
		t.Fatal("expected mr list without repo to fail")
	}
	if stderr == "" {
		t.Fatal("expected argument error on stderr")
	}
}

type fakeMergeRequestUseCase struct {
	list        []app.MergeRequestListItem
	detail      app.MergeRequestDetail
	created     app.MergeRequestMutationResult
	merged      app.MergeRequestMutationResult
	listCalls   int
	createInput app.CreateMergeRequestInput
	mergeInput  app.MergeMergeRequestInput
}

func (u *fakeMergeRequestUseCase) ListMergeRequests(ctx context.Context, repo string) ([]app.MergeRequestListItem, error) {
	u.listCalls++
	return u.list, nil
}

func (u *fakeMergeRequestUseCase) GetMergeRequest(ctx context.Context, repo, id string) (app.MergeRequestDetail, error) {
	return u.detail, nil
}

func (u *fakeMergeRequestUseCase) CreateMergeRequest(ctx context.Context, input app.CreateMergeRequestInput) (app.MergeRequestMutationResult, error) {
	u.createInput = input
	return u.created, nil
}

func (u *fakeMergeRequestUseCase) MergeMergeRequest(ctx context.Context, input app.MergeMergeRequestInput) (app.MergeRequestMutationResult, error) {
	u.mergeInput = input
	return u.merged, nil
}
