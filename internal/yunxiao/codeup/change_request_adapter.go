package codeup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

type ChangeRequestAdapter struct {
	client *yunxiao.Client
}

func NewChangeRequestAdapter(config yunxiao.ClientConfig) *ChangeRequestAdapter {
	return &ChangeRequestAdapter{client: yunxiao.NewClient(config)}
}

func (a *ChangeRequestAdapter) ListMergeRequests(ctx context.Context, repo string) ([]app.MergeRequestListItem, error) {
	query := url.Values{}
	query.Set("repositoryId", repo)
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.organizationChangeRequestsPath(), query, nil)
	if err != nil {
		return nil, err
	}
	return decodeChangeRequests(data)
}

func (a *ChangeRequestAdapter) GetMergeRequest(ctx context.Context, repo, id string) (app.MergeRequestDetail, error) {
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.changeRequestPath(repo, id), nil, nil)
	if err != nil {
		return app.MergeRequestDetail{}, err
	}
	return decodeChangeRequest(data)
}

func (a *ChangeRequestAdapter) CreateMergeRequest(ctx context.Context, input app.CreateMergeRequestInput) (app.MergeRequestDetail, error) {
	body, err := json.Marshal(changeRequestCreateRequest{
		SourceBranch: input.SourceBranch,
		TargetBranch: input.TargetBranch,
		Title:        input.Title,
	})
	if err != nil {
		return app.MergeRequestDetail{}, fmt.Errorf("encode change request: %w", err)
	}
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPost, paths.changeRequestsPath(input.Repo), nil, body)
	if err != nil {
		return app.MergeRequestDetail{}, err
	}
	return decodeChangeRequest(data)
}

func (a *ChangeRequestAdapter) MergeMergeRequest(ctx context.Context, repo, id string) (app.MergeRequestDetail, error) {
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPost, paths.changeRequestMergePath(repo, id), nil, []byte(`{}`))
	if err != nil {
		return app.MergeRequestDetail{}, err
	}
	return decodeChangeRequest(data)
}
