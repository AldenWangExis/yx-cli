package codeup

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

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

type changeRequestResponse struct {
	ID           int64  `json:"id"`
	LocalID      int64  `json:"localId"`
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

func decodeRepositories(data []byte) ([]app.RepositoryListItem, int, error) {
	var response []repositoryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, 0, fmt.Errorf("decode repositories: %w", err)
	}
	repos := make([]app.RepositoryListItem, 0, len(response))
	for _, repo := range response {
		repos = append(repos, app.RepositoryListItem{
			ID:   strconv.FormatInt(repo.ID, 10),
			Name: repo.Name,
			Path: repo.PathWithNamespace,
		})
	}
	return repos, len(response), nil
}

func decodeRepository(data []byte) (app.RepositoryDetail, error) {
	var response repositoryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.RepositoryDetail{}, fmt.Errorf("decode repository: %w", err)
	}
	return repositoryDetail(response), nil
}

func repositoryDetail(response repositoryResponse) app.RepositoryDetail {
	return app.RepositoryDetail{
		ID:            strconv.FormatInt(response.ID, 10),
		Name:          response.Name,
		Path:          response.PathWithNamespace,
		CloneURL:      response.SSHURLToRepo,
		WebURL:        response.WebURL,
		DefaultBranch: response.DefaultBranch,
	}
}

func decodeBranches(data []byte) ([]app.BranchListItem, error) {
	var response []branchResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode branches: %w", err)
	}
	branches := make([]app.BranchListItem, 0, len(response))
	for _, branch := range response {
		branches = append(branches, branchListItem(branch))
	}
	return branches, nil
}

func decodeBranch(data []byte) (app.BranchListItem, error) {
	var response branchResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.BranchListItem{}, fmt.Errorf("decode branch: %w", err)
	}
	return branchListItem(response), nil
}

func branchListItem(response branchResponse) app.BranchListItem {
	return app.BranchListItem{
		Name:      response.Name,
		Default:   response.DefaultBranch,
		Protected: response.Protected,
		CommitID:  response.Commit.ID,
		WebURL:    response.WebURL,
	}
}

func decodeCommits(data []byte) ([]app.CommitListItem, error) {
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

func decodeFile(data []byte) (app.RepositoryFile, error) {
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

func decodeChangeRequests(data []byte) ([]app.MergeRequestListItem, error) {
	var response []changeRequestResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode change requests: %w", err)
	}
	items := make([]app.MergeRequestListItem, 0, len(response))
	for _, mr := range response {
		items = append(items, app.MergeRequestListItem{
			ID:           formatChangeRequestID(firstNonZero(mr.ID, mr.LocalID)),
			Title:        mr.Title,
			State:        mr.State,
			SourceBranch: mr.SourceBranch,
			TargetBranch: mr.TargetBranch,
		})
	}
	return items, nil
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
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

func firstNonZero(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
