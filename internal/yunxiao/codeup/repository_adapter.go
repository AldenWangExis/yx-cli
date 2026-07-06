package codeup

import (
	"context"
	"encoding/json"
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
	paths := newCodeupPaths(a.client)
	for page := 1; ; page++ {
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("perPage", strconv.Itoa(a.perPage))
		data, err := a.client.DoJSON(ctx, http.MethodGet, paths.repositoriesPath(), query, nil)
		if err != nil {
			return nil, err
		}
		items, count, err := decodeRepositories(data)
		if err != nil {
			return nil, err
		}
		repos = append(repos, items...)
		if count < a.perPage {
			break
		}
	}
	return repos, nil
}

func (a *RepositoryAdapter) GetRepository(ctx context.Context, id string) (app.RepositoryDetail, error) {
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.repositoryPath(id), nil, nil)
	if err != nil {
		return app.RepositoryDetail{}, err
	}
	return decodeRepository(data)
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
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPost, paths.repositoriesPath(), query, body)
	if err != nil {
		return app.RepositoryDetail{}, err
	}
	return decodeRepository(data)
}

func (a *RepositoryAdapter) DeleteRepository(ctx context.Context, repo string) (app.RepositoryDetail, error) {
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodDelete, paths.repositoryPath(repo), nil, nil)
	if err != nil {
		return app.RepositoryDetail{}, err
	}
	if len(data) == 0 {
		return app.RepositoryDetail{ID: repo}, nil
	}
	deleted, err := decodeRepository(data)
	if err != nil {
		return app.RepositoryDetail{}, err
	}
	if deleted.ID == "" {
		deleted.ID = repo
	}
	return deleted, nil
}

func (a *RepositoryAdapter) ListRepositoryMembers(ctx context.Context, repo string) ([]app.RepositoryMember, error) {
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.repositoryMembersPath(repo), nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeRepositoryMembers(data)
}

func (a *RepositoryAdapter) AddRepositoryMember(ctx context.Context, input app.AddRepositoryMemberInput) (app.RepositoryMember, error) {
	accessLevel, err := strconv.Atoi(input.AccessLevel)
	if err != nil {
		return app.RepositoryMember{}, err
	}
	query := url.Values{}
	query.Set("userIds", input.UserID)
	query.Set("accessLevel", strconv.Itoa(accessLevel))
	if input.ExpiresAt != "" {
		query.Set("expiresAt", input.ExpiresAt)
	}
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPost, paths.repositoryMembersPath(input.Repo), query, nil)
	if err != nil {
		return app.RepositoryMember{}, err
	}
	member, err := decodeRepositoryMember(data)
	if err != nil {
		return app.RepositoryMember{}, err
	}
	if member.UserID == "" {
		member.UserID = input.UserID
	}
	if member.AccessLevel == 0 {
		member.AccessLevel = accessLevel
		member.Access = app.RepositoryAccessLevelName(accessLevel)
	}
	return member, nil
}

func (a *RepositoryAdapter) UpdateRepositoryMember(ctx context.Context, input app.UpdateRepositoryMemberInput) (app.RepositoryMember, error) {
	accessLevel, err := strconv.Atoi(input.AccessLevel)
	if err != nil {
		return app.RepositoryMember{}, err
	}
	query := url.Values{}
	query.Set("accessLevel", strconv.Itoa(accessLevel))
	if input.ExpiresAt != "" {
		query.Set("expiresAt", input.ExpiresAt)
	}
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPut, paths.repositoryMemberPath(input.Repo, input.UserID), query, nil)
	if err != nil {
		return app.RepositoryMember{}, err
	}
	member, err := decodeRepositoryMember(data)
	if err != nil {
		return app.RepositoryMember{}, err
	}
	if member.UserID == "" {
		member.UserID = input.UserID
	}
	if member.AccessLevel == 0 {
		member.AccessLevel = accessLevel
		member.Access = app.RepositoryAccessLevelName(accessLevel)
	}
	return member, nil
}

func (a *RepositoryAdapter) RemoveRepositoryMember(ctx context.Context, input app.RemoveRepositoryMemberInput) (app.RepositoryMember, error) {
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodDelete, paths.repositoryMemberPath(input.Repo, input.UserID), nil, nil)
	if err != nil {
		return app.RepositoryMember{}, err
	}
	member, err := decodeRepositoryMember(data)
	if err != nil {
		return app.RepositoryMember{}, err
	}
	if member.UserID == "" {
		member.UserID = input.UserID
	}
	return member, nil
}

func (a *RepositoryAdapter) ListBranches(ctx context.Context, repo string) ([]app.BranchListItem, error) {
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.repositoryBranchesPath(repo), nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeBranches(data)
}

func (a *RepositoryAdapter) SyncBranch(ctx context.Context, input app.BranchSyncInput) (app.BranchListItem, error) {
	query := url.Values{}
	query.Set("branch", input.Target)
	query.Set("ref", input.Source)
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPost, paths.repositoryBranchesPath(input.Repo), query, nil)
	if err != nil {
		return app.BranchListItem{}, err
	}
	return decodeBranch(data)
}

func (a *RepositoryAdapter) ListCommits(ctx context.Context, input app.CommitListInput) ([]app.CommitListItem, error) {
	query := url.Values{}
	query.Set("page", "1")
	query.Set("perPage", strconv.Itoa(a.perPage))
	if input.Ref != "" {
		query.Set("refName", input.Ref)
	}
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.repositoryCommitsPath(input.Repo), query, nil)
	if err != nil {
		return nil, err
	}
	return decodeCommits(data)
}

func (a *RepositoryAdapter) GetFile(ctx context.Context, input app.FileGetInput) (app.RepositoryFile, error) {
	query := url.Values{}
	query.Set("ref", input.Ref)
	paths := newCodeupPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.repositoryFilePath(input.Repo, input.Path), query, nil)
	if err != nil {
		return app.RepositoryFile{}, err
	}
	return decodeFile(data)
}
