package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
)

func TestAuthStatusLoginLogoutUseProvider(t *testing.T) {
	provider := &fakeAuthProvider{
		status: auth.Status{Profile: "default", HasToken: true, Backend: "fake", TokenMask: "pt-********c75c"},
	}
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Domain:       "https://devops.aliyun.com",
				Organization: "org-1",
				Region:       "center",
				ServiceConnections: map[string]string{
					"codeup": "sc-codeup-1",
					"flow":   "sc-flow-1",
				},
			},
		},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:          configPath,
		AuthProvider:        provider,
		AuthAccountResolver: fakeAuthAccountResolver{user: AuthAccount{Name: "王子豪", AccountID: "account-1"}},
		DefaultProfile:      "default",
	}), "auth", "status")
	if err != nil {
		t.Fatalf("expected auth status to succeed, got error: %v stderr=%s", err, stderr)
	}
	for _, want := range []string{
		"devops.aliyun.com",
		"✓ Logged in to devops.aliyun.com account 王子豪 (fake)",
		"- Active profile: true",
		"- Account status: authenticated",
		"- Account ID: account-1",
		"- Organization: org-1",
		"- Region: center",
		"- Token: pt-********c75c",
		"- Service connections:",
		"codeup: sc-****up-1",
		"flow: sc-**ow-1",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected status output to include %q, got:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "sc-codeup-1") || strings.Contains(stdout, "sc-flow-1") {
		t.Fatal("auth status leaked service connection id")
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

type fakeAuthAccountResolver struct {
	user AuthAccount
	err  error
}

func (r fakeAuthAccountResolver) ResolveAuthAccount(ctx context.Context, profileName string, profile config.Profile) (AuthAccount, error) {
	return r.user, r.err
}
