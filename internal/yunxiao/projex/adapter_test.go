package projex

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

func TestAdapterProjectsAndWorkitems(t *testing.T) {
	var createBody string
	var updateBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/projex/organizations/org-1/projects:search":
			_, _ = w.Write([]byte(`[{"id":"p1","name":"Project One"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems:search":
			_, _ = w.Write([]byte(`[{"id":"w1","subject":"Task One","status":{"name":"todo"},"workitemType":{"name":"task"},"space":{"id":"p1"}}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems/w1":
			_, _ = w.Write([]byte(`{"id":"w1","subject":"Task One","status":{"name":"todo"},"workitemType":{"name":"task"},"space":{"id":"p1"},"assignedTo":{"id":"u1"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/projex/organizations/org-1/workitems":
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			createBody = string(body)
			_, _ = w.Write([]byte(`{"id":"w2","subject":"Task Two","status":{"name":"todo"},"workitemType":{"name":"task"},"space":{"id":"p1"}}`))
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

	adapter := NewAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "token-1",
		OrganizationID: "org-1",
		Region:         "center",
	})

	projects, err := adapter.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("expected project list, got: %v", err)
	}
	if len(projects) != 1 || projects[0].ID != "p1" {
		t.Fatalf("unexpected projects: %+v", projects)
	}

	items, err := adapter.ListWorkitems(context.Background(), "p1")
	if err != nil {
		t.Fatalf("expected workitem list, got: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Task One" || items[0].ProjectID != "p1" {
		t.Fatalf("unexpected workitems: %+v", items)
	}

	detail, err := adapter.GetWorkitem(context.Background(), "w1")
	if err != nil {
		t.Fatalf("expected workitem detail, got: %v", err)
	}
	if detail.Assignee != "u1" {
		t.Fatalf("expected assignee u1, got %q", detail.Assignee)
	}

	created, err := adapter.CreateWorkitem(context.Background(), app.CreateWorkitemInput{ProjectID: "p1", Type: "task", Title: "Task Two"})
	if err != nil {
		t.Fatalf("expected create, got: %v", err)
	}
	if created.ID != "w2" {
		t.Fatalf("unexpected created item: %+v", created)
	}
	if !strings.Contains(createBody, `"spaceId":"p1"`) || !strings.Contains(createBody, `"subject":"Task Two"`) {
		t.Fatalf("unexpected create body: %s", createBody)
	}

	updated, err := adapter.UpdateWorkitem(context.Background(), app.UpdateWorkitemInput{ID: "w1", Status: "done", Assignee: "u2"})
	if err != nil {
		t.Fatalf("expected update, got: %v", err)
	}
	if updated.ID != "w1" || updated.Status != "done" || updated.Assignee != "u2" {
		t.Fatalf("unexpected updated item: %+v", updated)
	}
	if !strings.Contains(updateBody, `"status":"done"`) || !strings.Contains(updateBody, `"assignedTo":"u2"`) {
		t.Fatalf("unexpected update body: %s", updateBody)
	}
}

func TestAdapterDoesNotRetryNonIdempotentWorkitemWrite(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "temporary failure", http.StatusInternalServerError)
	}))
	defer server.Close()

	adapter := NewAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "token-1",
		OrganizationID: "org-1",
		Region:         "center",
	})

	_, err := adapter.CreateWorkitem(context.Background(), app.CreateWorkitemInput{ProjectID: "p1", Type: "task", Title: "Task Two"})
	if err == nil {
		t.Fatal("expected create to fail")
	}
	if calls != 1 {
		t.Fatalf("expected no retry for non-idempotent write, got %d calls", calls)
	}
}

func TestAdapterErrorsDoNotLeakToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "req-1")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"bad secret-token"}`))
	}))
	defer server.Close()

	adapter := NewAdapter(yunxiao.ClientConfig{
		BaseURL:        server.URL,
		Token:          "secret-token",
		OrganizationID: "org-1",
		Region:         "center",
	})

	_, err := adapter.ListProjects(context.Background())
	if err == nil {
		t.Fatal("expected API error")
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("error leaked token: %v", err)
	}
}
