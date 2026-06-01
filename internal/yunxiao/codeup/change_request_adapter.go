package codeup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.changeRequestsPath(repo), nil, nil)
	if err != nil {
		return nil, err
	}
	var response []changeRequestResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode change requests: %w", err)
	}
	items := make([]app.MergeRequestListItem, 0, len(response))
	for _, mr := range response {
		items = append(items, app.MergeRequestListItem{
			ID:           formatChangeRequestID(mr.ID),
			Title:        mr.Title,
			State:        mr.State,
			SourceBranch: mr.SourceBranch,
			TargetBranch: mr.TargetBranch,
		})
	}
	return items, nil
}

func (a *ChangeRequestAdapter) GetMergeRequest(ctx context.Context, repo, id string) (app.MergeRequestDetail, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.changeRequestPath(repo, id), nil, nil)
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
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.changeRequestsPath(input.Repo), nil, body)
	if err != nil {
		return app.MergeRequestDetail{}, err
	}
	return decodeChangeRequest(data)
}

func (a *ChangeRequestAdapter) MergeMergeRequest(ctx context.Context, repo, id string) (app.MergeRequestDetail, error) {
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.changeRequestPath(repo, id)+"/merge", nil, []byte(`{}`))
	if err != nil {
		return app.MergeRequestDetail{}, err
	}
	return decodeChangeRequest(data)
}

func (a *ChangeRequestAdapter) changeRequestsPath(repo string) string {
	if a.client.IsCenter() {
		return fmt.Sprintf("/oapi/v1/codeup/organizations/%s/repositories/%s/changeRequests", url.PathEscape(a.client.OrganizationID()), url.PathEscape(repo))
	}
	return fmt.Sprintf("/oapi/v1/codeup/repositories/%s/changeRequests", url.PathEscape(repo))
}

func (a *ChangeRequestAdapter) changeRequestPath(repo, id string) string {
	return a.changeRequestsPath(repo) + "/" + url.PathEscape(id)
}

func decodeChangeRequest(data []byte) (app.MergeRequestDetail, error) {
	var response changeRequestResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.MergeRequestDetail{}, fmt.Errorf("decode change request: %w", err)
	}
	return app.MergeRequestDetail{
		ID:           formatChangeRequestID(response.ID),
		Title:        response.Title,
		State:        response.State,
		SourceBranch: response.SourceBranch,
		TargetBranch: response.TargetBranch,
		WebURL:       response.WebURL,
	}, nil
}

func formatChangeRequestID(id int64) string {
	return strconv.FormatInt(id, 10)
}

type changeRequestResponse struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	State        string `json:"state"`
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	WebURL       string `json:"webUrl"`
}

type changeRequestCreateRequest struct {
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	Title        string `json:"title"`
}
