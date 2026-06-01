package app

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/safety"
)

type MergeRequestListItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	State        string `json:"state"`
	SourceBranch string `json:"sourceBranch,omitempty"`
	TargetBranch string `json:"targetBranch,omitempty"`
}

type MergeRequestDetail struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	State        string `json:"state"`
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	WebURL       string `json:"webUrl,omitempty"`
}

type CreateMergeRequestInput struct {
	Repo         string
	SourceBranch string
	TargetBranch string
	Title        string
	DryRun       bool
	Yes          bool
}

type MergeMergeRequestInput struct {
	Repo   string
	ID     string
	DryRun bool
	Yes    bool
}

type MergeRequestMutationResult struct {
	DryRun       bool               `json:"dryRun"`
	Summary      string             `json:"summary,omitempty"`
	MergeRequest MergeRequestDetail `json:"mergeRequest,omitempty"`
}

type MergeRequestService interface {
	ListMergeRequests(ctx context.Context, repo string) ([]MergeRequestListItem, error)
	GetMergeRequest(ctx context.Context, repo, id string) (MergeRequestDetail, error)
	CreateMergeRequest(ctx context.Context, input CreateMergeRequestInput) (MergeRequestDetail, error)
	MergeMergeRequest(ctx context.Context, repo, id string) (MergeRequestDetail, error)
}

type MergeRequestUseCase struct {
	service MergeRequestService
	safety  safety.Environment
}

func NewMergeRequestUseCase(service MergeRequestService, safetyEnv safety.Environment) *MergeRequestUseCase {
	return &MergeRequestUseCase{service: service, safety: safetyEnv}
}

func (u *MergeRequestUseCase) ListMergeRequests(ctx context.Context, repo string) ([]MergeRequestListItem, error) {
	return u.service.ListMergeRequests(ctx, repo)
}

func (u *MergeRequestUseCase) GetMergeRequest(ctx context.Context, repo, id string) (MergeRequestDetail, error) {
	return u.service.GetMergeRequest(ctx, repo, id)
}

func (u *MergeRequestUseCase) CreateMergeRequest(ctx context.Context, input CreateMergeRequestInput) (MergeRequestMutationResult, error) {
	if input.Repo == "" || input.SourceBranch == "" || input.TargetBranch == "" || input.Title == "" {
		return MergeRequestMutationResult{}, fmt.Errorf("repo, source branch, target branch, and title are required")
	}
	summary := fmt.Sprintf("create merge request %q from %s to %s in %s", input.Title, input.SourceBranch, input.TargetBranch, input.Repo)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return MergeRequestMutationResult{}, err
	}
	if decision.DryRun {
		return MergeRequestMutationResult{DryRun: true, Summary: summary}, nil
	}
	detail, err := u.service.CreateMergeRequest(ctx, input)
	if err != nil {
		return MergeRequestMutationResult{}, err
	}
	return MergeRequestMutationResult{MergeRequest: detail}, nil
}

func (u *MergeRequestUseCase) MergeMergeRequest(ctx context.Context, input MergeMergeRequestInput) (MergeRequestMutationResult, error) {
	if input.Repo == "" || input.ID == "" {
		return MergeRequestMutationResult{}, fmt.Errorf("repo and merge request id are required")
	}
	summary := fmt.Sprintf("merge merge request %s in %s", input.ID, input.Repo)
	decision, err := safety.Decide(safety.Request{Summary: summary, DryRun: input.DryRun, Yes: input.Yes}, u.safety)
	if err != nil {
		return MergeRequestMutationResult{}, err
	}
	if decision.DryRun {
		return MergeRequestMutationResult{DryRun: true, Summary: summary}, nil
	}
	detail, err := u.service.MergeMergeRequest(ctx, input.Repo, input.ID)
	if err != nil {
		return MergeRequestMutationResult{}, err
	}
	return MergeRequestMutationResult{MergeRequest: detail}, nil
}
