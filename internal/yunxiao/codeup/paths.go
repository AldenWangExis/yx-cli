package codeup

import (
	"fmt"
	"net/url"

	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

type codeupPaths struct {
	client *yunxiao.Client
}

func newCodeupPaths(client *yunxiao.Client) codeupPaths {
	return codeupPaths{client: client}
}

func (p codeupPaths) repositoriesPath() string {
	if p.client.IsCenter() {
		return fmt.Sprintf("/oapi/v1/codeup/organizations/%s/repositories", url.PathEscape(p.client.OrganizationID()))
	}
	return "/oapi/v1/codeup/repositories"
}

func (p codeupPaths) repositoryPath(id string) string {
	if p.client.IsCenter() {
		return fmt.Sprintf("/oapi/v1/codeup/organizations/%s/repositories/%s", url.PathEscape(p.client.OrganizationID()), url.PathEscape(id))
	}
	return fmt.Sprintf("/oapi/v1/codeup/repositories/%s", url.PathEscape(id))
}

func (p codeupPaths) repositoryBranchesPath(repo string) string {
	return p.repositoryPath(repo) + "/branches"
}

func (p codeupPaths) repositoryCommitsPath(repo string) string {
	return p.repositoryPath(repo) + "/commits"
}

func (p codeupPaths) repositoryFilePath(repo, path string) string {
	return p.repositoryPath(repo) + "/files/" + url.PathEscape(path)
}

func (p codeupPaths) changeRequestsPath(repo string) string {
	if p.client.IsCenter() {
		return fmt.Sprintf("/oapi/v1/codeup/organizations/%s/repositories/%s/changeRequests", url.PathEscape(p.client.OrganizationID()), url.PathEscape(repo))
	}
	return fmt.Sprintf("/oapi/v1/codeup/repositories/%s/changeRequests", url.PathEscape(repo))
}

func (p codeupPaths) organizationChangeRequestsPath() string {
	if p.client.IsCenter() {
		return fmt.Sprintf("/oapi/v1/codeup/organizations/%s/changeRequests", url.PathEscape(p.client.OrganizationID()))
	}
	return "/oapi/v1/codeup/changeRequests"
}

func (p codeupPaths) changeRequestPath(repo, id string) string {
	return p.changeRequestsPath(repo) + "/" + url.PathEscape(id)
}

func (p codeupPaths) changeRequestMergePath(repo, id string) string {
	return p.changeRequestPath(repo, id) + "/merge"
}

func (p codeupPaths) changeRequestClosePath(repo, id string) string {
	return p.changeRequestPath(repo, id) + "/close"
}
