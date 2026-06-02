package projex

import (
	"fmt"
	"net/url"

	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

type projexPaths struct {
	client *yunxiao.Client
}

func newProjexPaths(client *yunxiao.Client) projexPaths {
	return projexPaths{client: client}
}

func (p projexPaths) orgPath(path string) string {
	return fmt.Sprintf("/oapi/v1/projex/organizations/%s%s", url.PathEscape(p.client.OrganizationID()), path)
}

func (p projexPaths) projectsSearchPath() string {
	return p.orgPath("/projects:search")
}

func (p projexPaths) projectTemplatesPath() string {
	return p.orgPath("/projectTemplates")
}

func (p projexPaths) projectsPath() string {
	return p.orgPath("/projects")
}

func (p projexPaths) workitemsSearchPath() string {
	return p.orgPath("/workitems:search")
}

func (p projexPaths) workitemsPath() string {
	return p.orgPath("/workitems")
}

func (p projexPaths) workitemPath(id string) string {
	return p.orgPath("/workitems/" + url.PathEscape(id))
}
