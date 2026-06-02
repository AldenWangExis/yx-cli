package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
)

func TestRepoListBuildsDefaultUseCaseFromConfigAndToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"name":"demo","pathWithNamespace":"org/demo"}]`))
	}))
	defer server.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Domain:       server.URL,
				Organization: "org-1",
				Region:       "center",
			},
		},
	}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("default", "token-1"); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"--json", "repo", "list")
	if err != nil {
		t.Fatalf("expected repo list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var repos []app.RepositoryListItem
	if err := json.Unmarshal([]byte(stdout), &repos); err != nil {
		t.Fatalf("expected JSON repos, got error: %v output=%s", err, stdout)
	}
	if len(repos) != 1 || repos[0].Name != "demo" {
		t.Fatalf("unexpected repos: %+v", repos)
	}
}

func TestRepoListUsesCommandProfileDomainAndOrganizationOverrides(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-alt" {
			t.Fatalf("expected alt token, got %q", r.Header.Get("x-yunxiao-token"))
		}
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-override/repositories" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":2,"name":"override-demo","pathWithNamespace":"org-override/demo"}]`))
	}))
	defer server.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Domain:       server.URL,
				Organization: "org-default",
				Region:       "center",
			},
			"alt": {
				Domain:       "https://devops.aliyun.com",
				Organization: "org-alt",
				Region:       "center",
			},
		},
	}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	tokenStore := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml"))
	if err := tokenStore.Save("default", "token-default"); err != nil {
		t.Fatalf("failed to save default token: %v", err)
	}
	if err := tokenStore.Save("alt", "token-alt"); err != nil {
		t.Fatalf("failed to save alt token: %v", err)
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"--profile", "alt", "--domain", server.URL, "--org", "org-override", "--json", "repo", "list")
	if err != nil {
		t.Fatalf("expected repo list with overrides to succeed, got error: %v stderr=%s", err, stderr)
	}
	var repos []app.RepositoryListItem
	if err := json.Unmarshal([]byte(stdout), &repos); err != nil {
		t.Fatalf("expected JSON repos, got error: %v output=%s", err, stdout)
	}
	if len(repos) != 1 || repos[0].Name != "override-demo" {
		t.Fatalf("unexpected repos: %+v", repos)
	}
}

func TestRepoCurrentBuildsDefaultResolverFromConfigTokenGitRemoteAndCaches(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":6925918,"name":"yx-cli","pathWithNamespace":"org-1/yx-cli"}]`))
	}))
	defer server.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Domain:       server.URL,
				Organization: "org-1",
				Region:       "center",
			},
		},
	}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("default", "token-1"); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	repoDir := filepath.Join(dir, "repo")
	runGit(t, dir, "init", "repo")
	runGit(t, repoDir, "remote", "add", "origin", "git@codeup.aliyun.com:org-1/yx-cli.git")

	opts := Options{ConfigPath: configPath, WorkDir: repoDir}
	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts),
		"--json", "repo", "current", "--refresh")
	if err != nil {
		t.Fatalf("expected repo current to succeed, got error: %v stderr=%s", err, stderr)
	}
	var current app.CurrentRepository
	if err := json.Unmarshal([]byte(stdout), &current); err != nil {
		t.Fatalf("expected JSON current repo, got error: %v output=%s", err, stdout)
	}
	if current.ID != "6925918" || current.Path != "org-1/yx-cli" || current.Source != "api" {
		t.Fatalf("unexpected current repo: %+v", current)
	}

	loaded, err := config.NewStore(configPath).Load()
	if err != nil {
		t.Fatalf("expected config to load after cache write, got: %v", err)
	}
	var cachedIdentity config.RepoIdentity
	for _, identity := range loaded.Profiles["default"].RepoIdentityMap {
		if identity.Path == "org-1/yx-cli" {
			cachedIdentity = identity
			break
		}
	}
	if cachedIdentity.ID != "6925918" {
		t.Fatalf("expected repo identity cache, got %+v", loaded.Profiles["default"].RepoIdentityMap)
	}
	if cachedIdentity.Domain == "" || cachedIdentity.Organization != "org-1" || cachedIdentity.Region != "center" {
		t.Fatalf("expected scoped repo identity cache, got %+v", cachedIdentity)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts),
		"--json", "repo", "current")
	if err != nil {
		t.Fatalf("expected cached repo current to succeed, got error: %v stderr=%s", err, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &current); err != nil {
		t.Fatalf("expected JSON cached repo, got error: %v output=%s", err, stdout)
	}
	if current.ID != "6925918" || current.Source != "cache" {
		t.Fatalf("expected cached current repo, got %+v", current)
	}
	if requests != 1 {
		t.Fatalf("expected one API request because second call uses cache, got %d", requests)
	}
}

func TestRepoCurrentCacheIsScopedByDomain(t *testing.T) {
	var firstRequests int
	firstServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstRequests++
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories" {
			t.Fatalf("unexpected first path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":111,"name":"yx-cli","pathWithNamespace":"org-1/yx-cli"}]`))
	}))
	defer firstServer.Close()

	var secondRequests int
	secondServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondRequests++
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories" {
			t.Fatalf("unexpected second path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":222,"name":"yx-cli","pathWithNamespace":"org-1/yx-cli"}]`))
	}))
	defer secondServer.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	store := config.NewStore(configPath)
	if err := store.Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Domain:       firstServer.URL,
				Organization: "org-1",
				Region:       "center",
			},
		},
	}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("default", "token-1"); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	repoDir := filepath.Join(dir, "repo")
	runGit(t, dir, "init", "repo")
	runGit(t, repoDir, "remote", "add", "origin", "git@codeup.aliyun.com:org-1/yx-cli.git")

	opts := Options{ConfigPath: configPath, WorkDir: repoDir}
	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts),
		"--json", "repo", "current", "--refresh")
	if err != nil {
		t.Fatalf("expected first repo current to succeed, got error: %v stderr=%s", err, stderr)
	}
	var current app.CurrentRepository
	if err := json.Unmarshal([]byte(stdout), &current); err != nil {
		t.Fatalf("expected JSON current repo, got error: %v output=%s", err, stdout)
	}
	if current.ID != "111" || current.Source != "api" {
		t.Fatalf("unexpected first current repo: %+v", current)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("expected config load, got: %v", err)
	}
	profile := cfg.Profiles["default"]
	profile.Domain = secondServer.URL
	cfg.Profiles["default"] = profile
	if err := store.Save(cfg); err != nil {
		t.Fatalf("failed to switch domain: %v", err)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts),
		"--json", "repo", "current")
	if err != nil {
		t.Fatalf("expected second repo current to succeed, got error: %v stderr=%s", err, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &current); err != nil {
		t.Fatalf("expected JSON current repo, got error: %v output=%s", err, stdout)
	}
	if current.ID != "222" || current.Source != "api" {
		t.Fatalf("expected domain-scoped cache miss and second API result, got %+v", current)
	}
	if firstRequests != 1 || secondRequests != 1 {
		t.Fatalf("expected one request per domain, got first=%d second=%d", firstRequests, secondRequests)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
