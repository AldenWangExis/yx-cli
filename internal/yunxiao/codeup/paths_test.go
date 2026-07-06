package codeup

import (
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

func TestCodeupPathsShareCenterOrganizationRootAndEscapeIdentifiers(t *testing.T) {
	paths := newCodeupPaths(yunxiao.NewClient(yunxiao.ClientConfig{
		BaseURL:        "https://example.test",
		OrganizationID: "org/1",
		Region:         "center",
	}))

	if got, want := paths.repositoriesPath(), "/oapi/v1/codeup/organizations/org%2F1/repositories"; got != want {
		t.Fatalf("repositories path = %q, want %q", got, want)
	}
	if got, want := paths.repositoryPath("group/demo"), "/oapi/v1/codeup/organizations/org%2F1/repositories/group%2Fdemo"; got != want {
		t.Fatalf("repository path = %q, want %q", got, want)
	}
	if got, want := paths.repositoryMembersPath("group/demo"), "/oapi/v1/codeup/organizations/org%2F1/repositories/group%2Fdemo/members"; got != want {
		t.Fatalf("repository members path = %q, want %q", got, want)
	}
	if got, want := paths.repositoryMemberPath("group/demo", "user/1"), "/oapi/v1/codeup/organizations/org%2F1/repositories/group%2Fdemo/members/user%2F1"; got != want {
		t.Fatalf("repository member path = %q, want %q", got, want)
	}
	if got, want := paths.organizationChangeRequestsPath(), "/oapi/v1/codeup/organizations/org%2F1/changeRequests"; got != want {
		t.Fatalf("organization change requests path = %q, want %q", got, want)
	}
	if got, want := paths.changeRequestPath("group/demo", "mr/1"), "/oapi/v1/codeup/organizations/org%2F1/repositories/group%2Fdemo/changeRequests/mr%2F1"; got != want {
		t.Fatalf("change request path = %q, want %q", got, want)
	}
	if got, want := paths.changeRequestClosePath("group/demo", "mr/1"), "/oapi/v1/codeup/organizations/org%2F1/repositories/group%2Fdemo/changeRequests/mr%2F1/close"; got != want {
		t.Fatalf("change request close path = %q, want %q", got, want)
	}
}

func TestCodeupPathsUseRegionRoot(t *testing.T) {
	paths := newCodeupPaths(yunxiao.NewClient(yunxiao.ClientConfig{
		BaseURL: "https://example.test",
		Region:  "region",
	}))

	if got, want := paths.repositoriesPath(), "/oapi/v1/codeup/repositories"; got != want {
		t.Fatalf("repositories path = %q, want %q", got, want)
	}
	if got, want := paths.repositoryPath("group/demo"), "/oapi/v1/codeup/repositories/group%2Fdemo"; got != want {
		t.Fatalf("repository path = %q, want %q", got, want)
	}
	if got, want := paths.repositoryMembersPath("group/demo"), "/oapi/v1/codeup/repositories/group%2Fdemo/members"; got != want {
		t.Fatalf("repository members path = %q, want %q", got, want)
	}
	if got, want := paths.repositoryMemberPath("group/demo", "user/1"), "/oapi/v1/codeup/repositories/group%2Fdemo/members/user%2F1"; got != want {
		t.Fatalf("repository member path = %q, want %q", got, want)
	}
	if got, want := paths.organizationChangeRequestsPath(), "/oapi/v1/codeup/changeRequests"; got != want {
		t.Fatalf("organization change requests path = %q, want %q", got, want)
	}
	if got, want := paths.changeRequestPath("group/demo", "mr/1"), "/oapi/v1/codeup/repositories/group%2Fdemo/changeRequests/mr%2F1"; got != want {
		t.Fatalf("change request path = %q, want %q", got, want)
	}
	if got, want := paths.changeRequestClosePath("group/demo", "mr/1"), "/oapi/v1/codeup/repositories/group%2Fdemo/changeRequests/mr%2F1/close"; got != want {
		t.Fatalf("change request close path = %q, want %q", got, want)
	}
}
