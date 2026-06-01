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

func TestChangeRequestAdapterListGetCreateMerge(t *testing.T) {
	var createBody string
	var mergeCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/changeRequests" && r.URL.Query().Get("repositoryId") == "repo-1":
			_, _ = w.Write([]byte(`[{"id":11,"title":"Add feature","state":"opened","sourceBranch":"feat","targetBranch":"main"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/changeRequests/11":
			_, _ = w.Write([]byte(`{"id":11,"title":"Add feature","state":"opened","sourceBranch":"feat","targetBranch":"main","webUrl":"https://example.com/mr/11"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/changeRequests":
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			createBody = string(body)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":12,"title":"Add feature","state":"opened","sourceBranch":"feat","targetBranch":"main"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/codeup/organizations/org-1/repositories/repo-1/changeRequests/11/merge":
			mergeCalls++
			_, _ = w.Write([]byte(`{"id":11,"title":"Add feature","state":"merged","sourceBranch":"feat","targetBranch":"main"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewChangeRequestAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "token-1",
		OrganizationID: "org-1",
		Region:         "center",
	})

	listed, err := adapter.ListMergeRequests(context.Background(), "repo-1")
	if err != nil {
		t.Fatalf("expected list to succeed, got: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != "11" {
		t.Fatalf("unexpected listed MRs: %+v", listed)
	}

	detail, err := adapter.GetMergeRequest(context.Background(), "repo-1", "11")
	if err != nil {
		t.Fatalf("expected get to succeed, got: %v", err)
	}
	if detail.WebURL != "https://example.com/mr/11" {
		t.Fatalf("unexpected web URL: %q", detail.WebURL)
	}

	created, err := adapter.CreateMergeRequest(context.Background(), app.CreateMergeRequestInput{
		Repo:         "repo-1",
		SourceBranch: "feat",
		TargetBranch: "main",
		Title:        "Add feature",
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got: %v", err)
	}
	if created.ID != "12" {
		t.Fatalf("unexpected created MR: %+v", created)
	}
	if !strings.Contains(createBody, `"sourceBranch":"feat"`) || !strings.Contains(createBody, `"targetBranch":"main"`) {
		t.Fatalf("unexpected create body: %s", createBody)
	}

	merged, err := adapter.MergeMergeRequest(context.Background(), "repo-1", "11")
	if err != nil {
		t.Fatalf("expected merge to succeed, got: %v", err)
	}
	if merged.State != "merged" {
		t.Fatalf("expected merged state, got %q", merged.State)
	}
	if mergeCalls != 1 {
		t.Fatalf("expected one merge POST, got %d", mergeCalls)
	}
}

func TestChangeRequestAdapterDoesNotRetryNonIdempotentPost(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "temporary failure", http.StatusInternalServerError)
	}))
	defer server.Close()

	adapter := NewChangeRequestAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "token-1",
		OrganizationID: "org-1",
		Region:         "center",
	})

	_, err := adapter.CreateMergeRequest(context.Background(), app.CreateMergeRequestInput{
		Repo:         "repo-1",
		SourceBranch: "feat",
		TargetBranch: "main",
		Title:        "Add feature",
	})
	if err == nil {
		t.Fatal("expected create to fail")
	}
	if calls != 1 {
		t.Fatalf("expected no retry for non-idempotent POST, got %d calls", calls)
	}
}

func TestChangeRequestAdapterErrorsDoNotLeakToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "req-1")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"bad secret-token"}`))
	}))
	defer server.Close()

	adapter := NewChangeRequestAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "secret-token",
		OrganizationID: "org-1",
		Region:         "center",
	})

	_, err := adapter.ListMergeRequests(context.Background(), "repo-1")
	if err == nil {
		t.Fatal("expected API error")
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("error leaked token: %v", err)
	}
}
