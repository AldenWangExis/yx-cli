package codeup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
