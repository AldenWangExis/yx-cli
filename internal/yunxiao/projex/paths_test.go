package projex

import (
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

func TestProjexPathsUseOrganizationRootAndEscapeIdentifiers(t *testing.T) {
	paths := newProjexPaths(yunxiao.NewClient(yunxiao.ClientConfig{
		BaseURL:        "https://example.test",
		OrganizationID: "org/1",
		Region:         "center",
	}))

	if got, want := paths.projectsSearchPath(), "/oapi/v1/projex/organizations/org%2F1/projects:search"; got != want {
		t.Fatalf("projects search path = %q, want %q", got, want)
	}
	if got, want := paths.projectTemplatesPath(), "/oapi/v1/projex/organizations/org%2F1/projectTemplates"; got != want {
		t.Fatalf("project templates path = %q, want %q", got, want)
	}
	if got, want := paths.projectsPath(), "/oapi/v1/projex/organizations/org%2F1/projects"; got != want {
		t.Fatalf("projects path = %q, want %q", got, want)
	}
	if got, want := paths.workitemsSearchPath(), "/oapi/v1/projex/organizations/org%2F1/workitems:search"; got != want {
		t.Fatalf("workitems search path = %q, want %q", got, want)
	}
	if got, want := paths.workitemsPath(), "/oapi/v1/projex/organizations/org%2F1/workitems"; got != want {
		t.Fatalf("workitems path = %q, want %q", got, want)
	}
	if got, want := paths.workitemPath("wi/1"), "/oapi/v1/projex/organizations/org%2F1/workitems/wi%2F1"; got != want {
		t.Fatalf("workitem path = %q, want %q", got, want)
	}
}
