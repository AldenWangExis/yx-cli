package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/auth"
)

func TestAuthStatusLoginLogoutUseProvider(t *testing.T) {
	provider := &fakeAuthProvider{
		status: auth.Status{Profile: "default", HasToken: true, Backend: "fake"},
	}
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:     configPath,
		AuthProvider:   provider,
		DefaultProfile: "default",
	}), "auth", "status")
	if err != nil {
		t.Fatalf("expected auth status to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "default") || !strings.Contains(stdout, "fake") {
		t.Fatalf("expected status output to include profile and backend, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "secret-token") {
		t.Fatal("auth status leaked token")
	}

	cmd := NewRootCommandWithOptions(Options{
		ConfigPath:     configPath,
		AuthProvider:   provider,
		DefaultProfile: "default",
	})
	cmd.SetIn(strings.NewReader("secret-token\n"))
	stdout, stderr, err = executeCommand(t, cmd, "auth", "login")
	if err != nil {
		t.Fatalf("expected auth login to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "Yunxiao personal access token") {
		t.Fatalf("expected login prompt, got stdout:\n%s", stdout)
	}
	if provider.token != "secret-token" {
		t.Fatal("expected login to pass token to provider")
	}
	if strings.Contains(stdout, "secret-token") || strings.Contains(stderr, "secret-token") {
		t.Fatal("auth login leaked token")
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:     configPath,
		AuthProvider:   provider,
		DefaultProfile: "default",
	}), "auth", "logout")
	if err != nil {
		t.Fatalf("expected auth logout to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !provider.loggedOut {
		t.Fatal("expected provider logout to be called")
	}
}

type fakeAuthProvider struct {
	status    auth.Status
	token     string
	loggedOut bool
}

func (p *fakeAuthProvider) Login(ctx context.Context, profile, token string) (auth.Status, error) {
	p.token = token
	return auth.Status{Profile: profile, HasToken: token != "", Backend: "fake"}, nil
}

func (p *fakeAuthProvider) Status(ctx context.Context, profile string) (auth.Status, error) {
	p.status.Profile = profile
	return p.status, nil
}

func (p *fakeAuthProvider) Logout(ctx context.Context, profile string) error {
	p.loggedOut = true
	return nil
}
