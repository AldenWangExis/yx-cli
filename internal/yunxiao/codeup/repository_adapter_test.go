package codeup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

func TestRepositoryAdapterListRepositoriesCenterEndpoint(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("perPage") != "100" {
			t.Fatalf("unexpected pagination query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":2813489,"name":"demo","pathWithNamespace":"org/demo"}]`))
	}))
	defer server.Close()

	adapter := NewRepositoryAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "token-1",
		OrganizationID: "org-1",
		Region:         "center",
	})

	repos, err := adapter.ListRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected list to succeed, got: %v", err)
	}
	if requests != 1 {
		t.Fatalf("expected one request, got %d", requests)
	}
	if len(repos) != 1 || repos[0].ID != "2813489" || repos[0].Path != "org/demo" {
		t.Fatalf("unexpected repos: %+v", repos)
	}
}

func TestRepositoryAdapterListRepositoriesPaginatesUntilShortPage(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/oapi/v1/codeup/organizations/org-1/repositories" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("perPage") != "2" {
			t.Fatalf("unexpected perPage: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`[{"id":1,"name":"a","pathWithNamespace":"org/a"},{"id":2,"name":"b","pathWithNamespace":"org/b"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":3,"name":"c","pathWithNamespace":"org/c"}]`))
		default:
			t.Fatalf("unexpected page: %s", r.URL.Query().Get("page"))
		}
	}))
	defer server.Close()

	adapter := NewRepositoryAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "token-1",
		OrganizationID: "org-1",
		Region:         "center",
	})
	adapter.perPage = 2

	repos, err := adapter.ListRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected list to succeed, got: %v", err)
	}
	if requests != 2 {
		t.Fatalf("expected two paginated requests, got %d", requests)
	}
	if len(repos) != 3 || repos[2].Path != "org/c" {
		t.Fatalf("unexpected repositories: %+v", repos)
	}
}

func TestRepositoryAdapterGetRepositoryRegionEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/oapi/v1/codeup/repositories/org%2Fdemo" {
			t.Fatalf("unexpected path: %s", r.URL.EscapedPath())
		}
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":2813489,"name":"demo","pathWithNamespace":"org/demo","sshUrlToRepo":"git@example.com:org/demo.git"}`))
	}))
	defer server.Close()

	adapter := NewRepositoryAdapter(yunxiao.ClientConfig{
		BaseURL: server.URL,
		Token:   "token-1",
		Region:  "region",
	})

	repo, err := adapter.GetRepository(context.Background(), "org/demo")
	if err != nil {
		t.Fatalf("expected get to succeed, got: %v", err)
	}
	if repo.CloneURL != "git@example.com:org/demo.git" {
		t.Fatalf("expected ssh clone URL, got %q", repo.CloneURL)
	}
}

func TestRepositoryAdapterRepositoryOperations(t *testing.T) {
	var createBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories":
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			createBody = string(body)
			_, _ = w.Write([]byte(`{"id":2813490,"name":"created","pathWithNamespace":"org/created","sshUrlToRepo":"git@example.com:org/created.git","webUrl":"https://codeup.example/org/created"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/2813490/branches":
			_, _ = w.Write([]byte(`[{"name":"master","defaultBranch":true,"protected":false,"commit":{"id":"abc123","shortId":"abc123","title":"Initial commit"},"webUrl":"https://codeup.example/org/created/tree/master"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/2813490/branches":
			if r.URL.Query().Get("branch") != "feature/a" || r.URL.Query().Get("ref") != "master" {
				t.Fatalf("unexpected branch sync query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"name":"feature/a","commit":{"id":"abc123","shortId":"abc123","title":"Initial commit"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/2813490/commits":
			if r.URL.Query().Get("refName") != "master" {
				t.Fatalf("unexpected commit query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[{"id":"abc123","shortId":"abc123","title":"Initial commit","message":"Initial commit","authorName":"A","committedDate":"2026-06-01T10:00:00+08:00","webUrl":"https://codeup.example/commit/abc123"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/2813490/files/test.py":
			if r.URL.Query().Get("ref") != "master" {
				t.Fatalf("unexpected file query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"filePath":"test.py","ref":"master","encoding":"base64","content":"cHJpbnQoMSkK"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	adapter := NewRepositoryAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "token-1",
		OrganizationID: "org-1",
		Region:         "center",
	})

	created, err := adapter.CreateRepository(context.Background(), app.CreateRepositoryInput{Name: "created", Path: "created", Visibility: "private", ReadmeType: "EMPTY"})
	if err != nil {
		t.Fatalf("expected create repository, got: %v", err)
	}
	if created.ID != "2813490" || !strings.Contains(createBody, `"name":"created"`) {
		t.Fatalf("unexpected create result/body: %+v body=%s", created, createBody)
	}

	branches, err := adapter.ListBranches(context.Background(), "2813490")
	if err != nil {
		t.Fatalf("expected branches, got: %v", err)
	}
	if len(branches) != 1 || !branches[0].Default || branches[0].CommitID != "abc123" {
		t.Fatalf("unexpected branches: %+v", branches)
	}

	synced, err := adapter.SyncBranch(context.Background(), app.BranchSyncInput{Repo: "2813490", Source: "master", Target: "feature/a"})
	if err != nil {
		t.Fatalf("expected branch sync, got: %v", err)
	}
	if synced.Name != "feature/a" {
		t.Fatalf("unexpected synced branch: %+v", synced)
	}

	commits, err := adapter.ListCommits(context.Background(), app.CommitListInput{Repo: "2813490", Ref: "master"})
	if err != nil {
		t.Fatalf("expected commits, got: %v", err)
	}
	if len(commits) != 1 || commits[0].ShortID != "abc123" {
		t.Fatalf("unexpected commits: %+v", commits)
	}

	file, err := adapter.GetFile(context.Background(), app.FileGetInput{Repo: "2813490", Path: "test.py", Ref: "master"})
	if err != nil {
		t.Fatalf("expected file, got: %v", err)
	}
	if file.Content != "print(1)\n" {
		t.Fatalf("unexpected file: %+v", file)
	}
}

func TestRepositoryAdapterErrorsDoNotLeakToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "req-1")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"bad secret-token"}`))
	}))
	defer server.Close()

	adapter := NewRepositoryAdapter(yunxiao.ClientConfig{
		BaseURL: server.URL,
		Token:   "secret-token",
		Region:  "region",
	})

	_, err := adapter.ListRepositories(context.Background())
	if err == nil {
		t.Fatal("expected API error")
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("error leaked token: %v", err)
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "req-1") {
		t.Fatalf("expected status and request id in error, got: %v", err)
	}
}
