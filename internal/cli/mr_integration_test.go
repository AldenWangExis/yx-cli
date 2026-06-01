package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
)

func TestMRListBuildsDefaultUseCaseFromConfigAndToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/changeRequests" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":11,"title":"Add feature","state":"opened","sourceBranch":"feat","targetBranch":"main"}]`))
	}))
	defer server.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	saveYunxiaoProfile(t, configPath, server.URL)
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("default", "token-1"); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"--json", "mr", "list", "--repo", "repo-1")
	if err != nil {
		t.Fatalf("expected mr list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var mrs []app.MergeRequestListItem
	if err := json.Unmarshal([]byte(stdout), &mrs); err != nil {
		t.Fatalf("expected JSON MRs, got error: %v output=%s", err, stdout)
	}
	if len(mrs) != 1 || mrs[0].ID != "11" {
		t.Fatalf("unexpected MRs: %+v", mrs)
	}
}

func saveYunxiaoProfile(t *testing.T, configPath, domain string) {
	t.Helper()
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Domain:       domain,
				Organization: "org-1",
				Region:       "center",
			},
		},
	}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
}
