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

func TestMemberListBuildsDefaultUseCaseFromConfigAndToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		if r.Method != http.MethodGet || r.URL.Path != "/oapi/v1/platform/organizations/org-1/members" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"m1","userId":"u1","name":"王子豪","email":"wang@example.com","status":"ENABLED","roleIds":["admin"]}]`))
	}))
	defer server.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {Domain: server.URL, Organization: "org-1"},
		},
	}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("default", "token-1"); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"--json", "member", "list")
	if err != nil {
		t.Fatalf("expected member list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var members []app.Member
	if err := json.Unmarshal([]byte(stdout), &members); err != nil {
		t.Fatalf("expected JSON members, got error: %v output=%s", err, stdout)
	}
	if len(members) != 1 || members[0].UserID != "u1" {
		t.Fatalf("unexpected members: %+v", members)
	}
}

func TestWorkitemCreateAndUpdateResolveAssigneeMe(t *testing.T) {
	var createBody string
	var updateBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/platform/user":
			_, _ = w.Write([]byte(`{"id":"u-me","name":"王子豪"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems":
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			createBody = string(body)
			_, _ = w.Write([]byte(`{"id":"w1","subject":"Task One","status":{"name":"todo"},"workitemType":{"name":"task"},"space":{"id":"p1"},"assignedTo":{"id":"u-me"}}`))
		case r.Method == http.MethodPut && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems/w1":
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			updateBody = string(body)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {Domain: server.URL, Organization: "org-1"},
		},
	}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("default", "token-1"); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"issue", "create", "--project", "p1", "--type", "task", "--title", "Task One", "--assignee", "@me", "--yes")
	if err != nil {
		t.Fatalf("expected issue create to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(createBody, `"assignedTo":"u-me"`) {
		t.Fatalf("expected @me to resolve in create body, got %s", createBody)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"issue", "update", "w1", "--title", "P1 Task One", "--assignee", "@me", "--yes")
	if err != nil {
		t.Fatalf("expected issue update to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(updateBody, `"assignedTo":"u-me"`) || !strings.Contains(updateBody, `"subject":"P1 Task One"`) {
		t.Fatalf("expected @me and title in update body, got %s", updateBody)
	}
}
