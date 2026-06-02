package platform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

func TestCurrentUserUsesPlatformEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/oapi/v1/platform/user" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("expected token header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"account-1","name":"王子豪","username":"alden","email":"alden@example.com"}}`))
	}))
	defer server.Close()

	user, err := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1"}).CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("current user: %v", err)
	}
	if user.Name != "王子豪" || user.Username != "alden" || user.AccountID != "account-1" || user.Email != "alden@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestListOrganizationsUsesPlatformEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/oapi/v1/platform/organizations" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("perPage") != "100" {
			t.Fatalf("expected page query, got %s", r.URL.RawQuery)
		}
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("expected token header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id":"org-1","name":"研发组织","description":"Main org","createdAt":"2026-06-01T00:00:00Z"}
		]`))
	}))
	defer server.Close()

	organizations, err := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1"}).ListOrganizations(context.Background())
	if err != nil {
		t.Fatalf("list organizations: %v", err)
	}
	if len(organizations) != 1 {
		t.Fatalf("expected one organization, got %d", len(organizations))
	}
	got := organizations[0]
	if got.ID != "org-1" || got.Name != "研发组织" || got.Description != "Main org" || got.CreatedAt != "2026-06-01T00:00:00Z" {
		t.Fatalf("unexpected organization: %+v", got)
	}
}
