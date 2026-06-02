package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
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
	if !strings.Contains(stdout, "✓ Logged in profile default using fake token store") {
		t.Fatalf("expected explicit login success, got stdout:\n%s", stdout)
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

func TestAuthStatusShowsSetupHintsWhenNotLoggedIn(t *testing.T) {
	provider := &fakeAuthProvider{
		status: auth.Status{Profile: "default", HasToken: false, Backend: "fake"},
	}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:          filepath.Join(t.TempDir(), "config.yaml"),
		AuthProvider:        provider,
		AuthAccountResolver: fakeAuthAccountResolver{},
		DefaultProfile:      "default",
	}), "auth", "status")
	if err != nil {
		t.Fatalf("expected auth status to succeed, got error: %v stderr=%s", err, stderr)
	}

	for _, want := range []string{
		"Personal access token: https://account-devops.aliyun.com/settings/personalAccessToken",
		"Service connections: https://flow.aliyun.com/setting/service-connection",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected status output to include %q, got:\n%s", want, stdout)
		}
	}
}

func TestAuthLoginStoresOptionalServiceConnection(t *testing.T) {
	provider := &fakeAuthProvider{}
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	cmd := NewRootCommandWithOptions(Options{
		ConfigPath:     configPath,
		AuthProvider:   provider,
		DefaultProfile: "default",
	})
	cmd.SetIn(strings.NewReader("secret-token\nsc-codeup-1\n"))
	stdout, stderr, err := executeCommand(t, cmd, "auth", "login")
	if err != nil {
		t.Fatalf("expected auth login to succeed, got error: %v stderr=%s", err, stderr)
	}

	for _, want := range []string{
		"Yunxiao personal access token",
		"https://account-devops.aliyun.com/settings/personalAccessToken",
		"Codeup service connection ID",
		"https://flow.aliyun.com/setting/service-connection",
		"stored Codeup service connection for profile default",
		"✓ Logged in profile default using fake token store",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected login output to include %q, got:\n%s", want, stdout)
		}
	}
	if provider.token != "secret-token" {
		t.Fatal("expected login to pass token to provider")
	}
	cfg, err := config.NewStore(configPath).Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := cfg.Profiles["default"].ServiceConnections["codeup"]; got != "sc-codeup-1" {
		t.Fatalf("expected codeup service connection to be stored, got %q", got)
	}
	if strings.Contains(stdout, "secret-token") || strings.Contains(stderr, "secret-token") {
		t.Fatal("auth login leaked token")
	}
}

func TestAuthLoginAutoStoresSingleDiscoveredOrganization(t *testing.T) {
	provider := &fakeAuthProvider{}
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	usecase := &fakeOrganizationUseCase{organizations: []app.Organization{
		{ID: "org-1", Name: "研发组织"},
	}}

	cmd := NewRootCommandWithOptions(Options{
		ConfigPath:          configPath,
		AuthProvider:        provider,
		DefaultProfile:      "default",
		OrganizationUseCase: usecase,
	})
	cmd.SetIn(strings.NewReader("secret-token\n\n"))
	stdout, stderr, err := executeCommand(t, cmd, "auth", "login")
	if err != nil {
		t.Fatalf("expected auth login to succeed, got error: %v stderr=%s", err, stderr)
	}

	if !strings.Contains(stdout, "stored organization org-1 (研发组织) for profile default") {
		t.Fatalf("expected login to report stored organization, got:\n%s", stdout)
	}
	cfg, err := config.NewStore(configPath).Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	profile := cfg.Profiles["default"]
	if profile.Domain != "https://devops.aliyun.com" {
		t.Fatalf("expected login to default profile domain, got %q", profile.Domain)
	}
	if profile.Organization != "org-1" {
		t.Fatalf("expected login to store discovered organization, got %q", profile.Organization)
	}
	if usecase.listCalls != 1 {
		t.Fatalf("expected organization lookup once, got %d", usecase.listCalls)
	}
}

func TestAuthLoginPromptsWhenMultipleOrganizationsAreDiscovered(t *testing.T) {
	provider := &fakeAuthProvider{}
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	usecase := &fakeOrganizationUseCase{organizations: []app.Organization{
		{ID: "org-1", Name: "研发组织"},
		{ID: "org-2", Name: "测试组织"},
	}}

	cmd := NewRootCommandWithOptions(Options{
		ConfigPath:          configPath,
		AuthProvider:        provider,
		DefaultProfile:      "default",
		OrganizationUseCase: usecase,
	})
	cmd.SetIn(strings.NewReader("secret-token\n\norg-2\n"))
	stdout, stderr, err := executeCommand(t, cmd, "auth", "login")
	if err != nil {
		t.Fatalf("expected auth login to succeed, got error: %v stderr=%s", err, stderr)
	}

	for _, want := range []string{
		"org-1",
		"研发组织",
		"org-2",
		"测试组织",
		"Yunxiao organization ID (optional):",
		"stored organization org-2 for profile default",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected login output to include %q, got:\n%s", want, stdout)
		}
	}
	cfg, err := config.NewStore(configPath).Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Profiles["default"].Organization != "org-2" {
		t.Fatalf("expected selected organization to be saved, got %q", cfg.Profiles["default"].Organization)
	}
}

func TestAuthLoginSkipsEmptyServiceConnection(t *testing.T) {
	provider := &fakeAuthProvider{}
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	cmd := NewRootCommandWithOptions(Options{
		ConfigPath:     configPath,
		AuthProvider:   provider,
		DefaultProfile: "default",
	})
	cmd.SetIn(strings.NewReader("secret-token\n\n"))
	_, stderr, err := executeCommand(t, cmd, "auth", "login")
	if err != nil {
		t.Fatalf("expected auth login to succeed, got error: %v stderr=%s", err, stderr)
	}

	cfg, err := config.NewStore(configPath).Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := cfg.Profiles["default"].ServiceConnections["codeup"]; got != "" {
		t.Fatalf("expected empty service connection to be skipped, got %q", got)
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
