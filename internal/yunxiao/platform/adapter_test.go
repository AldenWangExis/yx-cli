package platform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
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

func TestMemberEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("expected token header")
		}
		w.WriteHeader(http.StatusOK)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/platform/organizations/org-1/members":
			if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("perPage") != "100" {
				t.Fatalf("expected paging query, got %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[{"id":"m1","userId":"u1","name":"王子豪","email":"wang@example.com","status":"ENABLED","roleIds":["admin"]}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/platform/organizations/org-1/members:search":
			_, _ = w.Write([]byte(`[{"id":"m2","userId":"u2","name":"王小明","status":"ENABLED"},{"id":"m3","userId":"u3","name":"李雷","status":"ENABLED"}]`))
		default:
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	}))
	defer server.Close()

	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1"})
	members, err := adapter.ListMembers(context.Background(), app.MemberListInput{Status: "ENABLED"})
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) != 1 || members[0].UserID != "u1" || members[0].RoleIDs[0] != "admin" {
		t.Fatalf("unexpected members: %+v", members)
	}

	searched, err := adapter.SearchMembers(context.Background(), app.MemberSearchInput{Name: "王"})
	if err != nil {
		t.Fatalf("search members: %v", err)
	}
	if len(searched) != 1 || searched[0].UserID != "u2" {
		t.Fatalf("unexpected searched members: %+v", searched)
	}
}
