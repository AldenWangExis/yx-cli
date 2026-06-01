package app

import (
	"bytes"
	"context"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

func TestMergeRequestUseCaseListsAndViews(t *testing.T) {
	service := &fakeMergeRequestService{
		list:   []MergeRequestListItem{{ID: "1", Title: "Add feature", State: "opened"}},
		detail: MergeRequestDetail{ID: "1", Title: "Add feature", SourceBranch: "feat", TargetBranch: "main"},
	}
	useCase := NewMergeRequestUseCase(service, safety.Environment{})

	listed, err := useCase.ListMergeRequests(context.Background(), "repo-1")
	if err != nil {
		t.Fatalf("expected list to succeed, got: %v", err)
	}
	if len(listed) != 1 || listed[0].Title != "Add feature" {
		t.Fatalf("unexpected merge requests: %+v", listed)
	}

	detail, err := useCase.GetMergeRequest(context.Background(), "repo-1", "1")
	if err != nil {
		t.Fatalf("expected get to succeed, got: %v", err)
	}
	if detail.SourceBranch != "feat" {
		t.Fatalf("expected source branch feat, got %q", detail.SourceBranch)
	}
}

func TestMergeRequestUseCaseCreateDryRunDoesNotMutate(t *testing.T) {
	service := &fakeMergeRequestService{}
	useCase := NewMergeRequestUseCase(service, safety.Environment{ConfirmWrites: true})

	result, err := useCase.CreateMergeRequest(context.Background(), CreateMergeRequestInput{
		Repo:         "repo-1",
		SourceBranch: "feat",
		TargetBranch: "main",
		Title:        "Add feature",
		DryRun:       true,
	})
	if err != nil {
		t.Fatalf("expected dry run to succeed, got: %v", err)
	}
	if !result.DryRun {
		t.Fatal("expected dry run result")
	}
	if service.createCalled {
		t.Fatal("expected dry run not to call create")
	}
}

func TestMergeRequestUseCaseMergeRequiresConfirmation(t *testing.T) {
	service := &fakeMergeRequestService{}
	useCase := NewMergeRequestUseCase(service, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	_, err := useCase.MergeMergeRequest(context.Background(), MergeMergeRequestInput{
		Repo: "repo-1",
		ID:   "1",
	})
	if err == nil {
		t.Fatal("expected merge to require confirmation")
	}
	if service.mergeCalled {
		t.Fatal("expected merge not to be called without confirmation")
	}
}

func TestMergeRequestUseCaseMergeWithYesMutatesOnce(t *testing.T) {
	service := &fakeMergeRequestService{
		merged: MergeRequestDetail{ID: "1", Title: "Add feature", State: "merged"},
	}
	useCase := NewMergeRequestUseCase(service, safety.Environment{ConfirmWrites: true, IsTerminal: false})

	result, err := useCase.MergeMergeRequest(context.Background(), MergeMergeRequestInput{
		Repo: "repo-1",
		ID:   "1",
		Yes:  true,
	})
	if err != nil {
		t.Fatalf("expected merge with yes to succeed, got: %v", err)
	}
	if result.MergeRequest.State != "merged" {
		t.Fatalf("expected merged state, got %q", result.MergeRequest.State)
	}
	if service.mergeCalls != 1 {
		t.Fatalf("expected one merge call, got %d", service.mergeCalls)
	}
}

func TestMergeRequestUseCaseInteractiveConfirmation(t *testing.T) {
	service := &fakeMergeRequestService{
		created: MergeRequestDetail{ID: "1", Title: "Add feature", State: "opened"},
	}
	var prompt bytes.Buffer
	useCase := NewMergeRequestUseCase(service, safety.Environment{
		ConfirmWrites: true,
		IsTerminal:    true,
		Input:         bytes.NewBufferString("yes\n"),
		Output:        &prompt,
	})

	_, err := useCase.CreateMergeRequest(context.Background(), CreateMergeRequestInput{
		Repo:         "repo-1",
		SourceBranch: "feat",
		TargetBranch: "main",
		Title:        "Add feature",
	})
	if err != nil {
		t.Fatalf("expected interactive create to succeed, got: %v", err)
	}
	if !service.createCalled {
		t.Fatal("expected create to be called after confirmation")
	}
	if prompt.String() == "" {
		t.Fatal("expected confirmation prompt")
	}
}

type fakeMergeRequestService struct {
	list         []MergeRequestListItem
	detail       MergeRequestDetail
	created      MergeRequestDetail
	merged       MergeRequestDetail
	createCalled bool
	mergeCalled  bool
	mergeCalls   int
}

func (s *fakeMergeRequestService) ListMergeRequests(ctx context.Context, repo string) ([]MergeRequestListItem, error) {
	return s.list, nil
}

func (s *fakeMergeRequestService) GetMergeRequest(ctx context.Context, repo, id string) (MergeRequestDetail, error) {
	return s.detail, nil
}

func (s *fakeMergeRequestService) CreateMergeRequest(ctx context.Context, input CreateMergeRequestInput) (MergeRequestDetail, error) {
	s.createCalled = true
	if s.created.ID == "" {
		return MergeRequestDetail{ID: "1", Title: input.Title, State: "opened"}, nil
	}
	return s.created, nil
}

func (s *fakeMergeRequestService) MergeMergeRequest(ctx context.Context, repo, id string) (MergeRequestDetail, error) {
	s.mergeCalled = true
	s.mergeCalls++
	if s.merged.ID == "" {
		return MergeRequestDetail{ID: id, State: "merged"}, nil
	}
	return s.merged, nil
}
