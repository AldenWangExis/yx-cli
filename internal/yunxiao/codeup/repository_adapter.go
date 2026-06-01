package codeup

import (
	"context"
	"encoding/base64"
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
	repos := []app.RepositoryListItem{}
	for page := 1; ; page++ {
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("perPage", strconv.Itoa(a.perPage))
		data, err := a.client.DoJSON(ctx, http.MethodGet, a.repositoriesPath(), query, nil)
		if err != nil {
			return nil, err
		}
		var response []repositoryResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("decode repositories: %w", err)
		}
		for _, repo := range response {
			repos = append(repos, app.RepositoryListItem{
				ID:   strconv.FormatInt(repo.ID, 10),
				Name: repo.Name,
				Path: repo.PathWithNamespace,
			})
		}
		if len(response) < a.perPage {
			break
		}
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
		ID:            strconv.FormatInt(response.ID, 10),
		Name:          response.Name,
		Path:          response.PathWithNamespace,
		CloneURL:      response.SSHURLToRepo,
		WebURL:        response.WebURL,
		DefaultBranch: response.DefaultBranch,
	}, nil
}

func (a *RepositoryAdapter) CreateRepository(ctx context.Context, input app.CreateRepositoryInput) (app.RepositoryDetail, error) {
	body, err := json.Marshal(repositoryCreateRequest{
		Name:        input.Name,
		Path:        input.Path,
		Description: input.Description,
		Visibility:  input.Visibility,
		ReadmeType:  input.ReadmeType,
	})
	if err != nil {
		return app.RepositoryDetail{}, err
	}
	query := url.Values{}
	query.Set("createParentPath", "true")
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.repositoriesPath(), query, body)
	if err != nil {
		return app.RepositoryDetail{}, err
	}
	var response repositoryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.RepositoryDetail{}, fmt.Errorf("decode repository: %w", err)
	}
	return app.RepositoryDetail{
		ID:            strconv.FormatInt(response.ID, 10),
		Name:          response.Name,
		Path:          response.PathWithNamespace,
		CloneURL:      response.SSHURLToRepo,
		WebURL:        response.WebURL,
		DefaultBranch: response.DefaultBranch,
	}, nil
}

func (a *RepositoryAdapter) ListBranches(ctx context.Context, repo string) ([]app.BranchListItem, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.repositoryPath(repo)+"/branches", nil, nil)
	if err != nil {
		return nil, err
	}
	var response []branchResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode branches: %w", err)
	}
	branches := make([]app.BranchListItem, 0, len(response))
	for _, branch := range response {
		branches = append(branches, decodeBranch(branch))
	}
	return branches, nil
}

func (a *RepositoryAdapter) SyncBranch(ctx context.Context, input app.BranchSyncInput) (app.BranchListItem, error) {
	query := url.Values{}
	query.Set("branch", input.Target)
	query.Set("ref", input.Source)
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.repositoryPath(input.Repo)+"/branches", query, nil)
	if err != nil {
		return app.BranchListItem{}, err
	}
	var response branchResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.BranchListItem{}, fmt.Errorf("decode branch: %w", err)
	}
	return decodeBranch(response), nil
}

func (a *RepositoryAdapter) ListCommits(ctx context.Context, input app.CommitListInput) ([]app.CommitListItem, error) {
	query := url.Values{}
	query.Set("page", "1")
	query.Set("perPage", strconv.Itoa(a.perPage))
	if input.Ref != "" {
		query.Set("refName", input.Ref)
	}
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.repositoryPath(input.Repo)+"/commits", query, nil)
	if err != nil {
		return nil, err
	}
	var response []commitResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode commits: %w", err)
	}
	commits := make([]app.CommitListItem, 0, len(response))
	for _, commit := range response {
		commits = append(commits, app.CommitListItem{
			ID:            commit.ID,
			ShortID:       commit.ShortID,
			Title:         commit.Title,
			Message:       commit.Message,
			AuthorName:    commit.AuthorName,
			CommittedDate: commit.CommittedDate,
			WebURL:        commit.WebURL,
		})
	}
	return commits, nil
}

func (a *RepositoryAdapter) GetFile(ctx context.Context, input app.FileGetInput) (app.RepositoryFile, error) {
	query := url.Values{}
	query.Set("ref", input.Ref)
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.repositoryPath(input.Repo)+"/files/"+url.PathEscape(input.Path), query, nil)
	if err != nil {
		return app.RepositoryFile{}, err
	}
	var response fileResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.RepositoryFile{}, fmt.Errorf("decode file: %w", err)
	}
	content := response.Content
	if response.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(response.Content)
		if err != nil {
			return app.RepositoryFile{}, fmt.Errorf("decode file content: %w", err)
		}
		content = string(decoded)
	}
	return app.RepositoryFile{
		Path:     response.FilePath,
		Ref:      response.Ref,
		Encoding: response.Encoding,
		Content:  content,
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
	HTTPURLToRepo     string `json:"httpUrlToRepo"`
	WebURL            string `json:"webUrl"`
	DefaultBranch     string `json:"defaultBranch"`
}

type repositoryCreateRequest struct {
	Name        string `json:"name"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	ReadmeType  string `json:"readMeType,omitempty"`
}

type branchResponse struct {
	Name          string         `json:"name"`
	DefaultBranch bool           `json:"defaultBranch"`
	Protected     bool           `json:"protected"`
	Commit        commitResponse `json:"commit"`
	WebURL        string         `json:"webUrl"`
}

type commitResponse struct {
	ID            string `json:"id"`
	ShortID       string `json:"shortId"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	AuthorName    string `json:"authorName"`
	CommittedDate string `json:"committedDate"`
	WebURL        string `json:"webUrl"`
}

type fileResponse struct {
	FilePath string `json:"filePath"`
	Ref      string `json:"ref"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

func decodeBranch(response branchResponse) app.BranchListItem {
	return app.BranchListItem{
		Name:      response.Name,
		Default:   response.DefaultBranch,
		Protected: response.Protected,
		CommitID:  response.Commit.ID,
		WebURL:    response.WebURL,
	}
}
