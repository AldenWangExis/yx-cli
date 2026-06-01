package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
)

func TestWorkitemListBuildsDefaultUseCaseFromConfigAndToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		if r.URL.Path != "/oapi/v1/projex/organizations/org-1/workitems:search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Body != nil {
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			if !strings.Contains(string(body), `"category":"Task"`) {
				_, _ = w.Write([]byte(`[]`))
				return
			}
		}
		_, _ = w.Write([]byte(`[{"id":"w1","subject":"Task One","status":{"name":"todo"},"workitemType":{"name":"task"},"space":{"id":"p1"}}]`))
	}))
	defer server.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Domain:         server.URL,
				Organization:   "org-1",
				Region:         "center",
				RepoProjectMap: map[string]string{"repo-a": "p1"},
			},
		},
	}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("default", "token-1"); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"--json", "issue", "list", "--repo", "repo-a")
	if err != nil {
		t.Fatalf("expected issue list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var items []app.WorkitemListItem
	if err := json.Unmarshal([]byte(stdout), &items); err != nil {
		t.Fatalf("expected JSON workitems, got error: %v output=%s", err, stdout)
	}
	if len(items) != 1 || items[0].ProjectID != "p1" {
		t.Fatalf("unexpected workitems: %+v", items)
	}
}
