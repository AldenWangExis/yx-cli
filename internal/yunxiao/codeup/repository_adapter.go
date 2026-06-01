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

type RepositoryAdapter struct {
	client  *yunxiao.Client
	perPage int
}

func NewRepositoryAdapter(config yunxiao.ClientConfig) *RepositoryAdapter {
	return &RepositoryAdapter{
		client:  yunxiao.NewClient(config),
		perPage: 100,
	}
}

func (a *RepositoryAdapter) ListRepositories(ctx context.Context) ([]app.RepositoryListItem, error) {
	query := url.Values{}
	query.Set("page", "1")
	query.Set("perPage", strconv.Itoa(a.perPage))
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.repositoriesPath(), query, nil)
	if err != nil {
		return nil, err
	}
	var response []repositoryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode repositories: %w", err)
	}
	repos := make([]app.RepositoryListItem, 0, len(response))
	for _, repo := range response {
		repos = append(repos, app.RepositoryListItem{
			ID:   strconv.FormatInt(repo.ID, 10),
			Name: repo.Name,
			Path: repo.PathWithNamespace,
		})
	}
	return repos, nil
}

func (a *RepositoryAdapter) GetRepository(ctx context.Context, id string) (app.RepositoryDetail, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.repositoryPath(id), nil, nil)
	if err != nil {
		return app.RepositoryDetail{}, err
	}
	var response repositoryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.RepositoryDetail{}, fmt.Errorf("decode repository: %w", err)
	}
	return app.RepositoryDetail{
		ID:       strconv.FormatInt(response.ID, 10),
		Name:     response.Name,
		Path:     response.PathWithNamespace,
		CloneURL: response.SSHURLToRepo,
	}, nil
}

func (a *RepositoryAdapter) repositoriesPath() string {
	if a.client.IsCenter() {
		return fmt.Sprintf("/oapi/v1/codeup/organizations/%s/repositories", url.PathEscape(a.client.OrganizationID()))
	}
	return "/oapi/v1/codeup/repositories"
}

func (a *RepositoryAdapter) repositoryPath(id string) string {
	if a.client.IsCenter() {
		return fmt.Sprintf("/oapi/v1/codeup/organizations/%s/repositories/%s", url.PathEscape(a.client.OrganizationID()), url.PathEscape(id))
	}
	return fmt.Sprintf("/oapi/v1/codeup/repositories/%s", url.PathEscape(id))
}

type repositoryResponse struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"pathWithNamespace"`
	SSHURLToRepo      string `json:"sshUrlToRepo"`
}
