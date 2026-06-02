package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/config"
)

func TestOrgListAndUse(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {Domain: "https://devops.aliyun.com"},
		},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	usecase := &fakeOrganizationUseCase{organizations: []app.Organization{
		{ID: "org-1", Name: "研发组织"},
		{ID: "org-2", Name: "测试组织"},
	}}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:          configPath,
		DefaultProfile:      "default",
		OrganizationUseCase: usecase,
		AuthAccountResolver: fakeAuthAccountResolver{},
	}), "--json", "org", "list")
	if err != nil {
		t.Fatalf("expected org list to succeed, got error: %v stderr=%s", err, stderr)
	}
	var listed []app.Organization
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatalf("expected JSON org list, got %v output=%s", err, stdout)
	}
	if len(listed) != 2 || listed[0].ID != "org-1" || listed[1].Name != "测试组织" {
		t.Fatalf("unexpected org list: %+v", listed)
	}
	if usecase.listCalls != 1 {
		t.Fatalf("expected org usecase to be called once, got %d", usecase.listCalls)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(Options{
		ConfigPath:          configPath,
		DefaultProfile:      "default",
		OrganizationUseCase: usecase,
	}), "org", "use", "org-2")
	if err != nil {
		t.Fatalf("expected org use to succeed, got error: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "set profile default organization to org-2") {
		t.Fatalf("expected org use confirmation, got:\n%s", stdout)
	}
	cfg, err := config.NewStore(configPath).Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Profiles["default"].Organization != "org-2" {
		t.Fatalf("expected organization to be saved, got %q", cfg.Profiles["default"].Organization)
	}
}

type fakeOrganizationUseCase struct {
	organizations []app.Organization
	listCalls     int
	err           error
}

func (u *fakeOrganizationUseCase) ListOrganizations(ctx context.Context) ([]app.Organization, error) {
	u.listCalls++
	if u.err != nil {
		return nil, u.err
	}
	return u.organizations, nil
}
