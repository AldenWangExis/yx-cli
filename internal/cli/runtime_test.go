package cli

import (
	"path/filepath"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
)

func TestResolveRuntimeProfileAppliesCommandContextOverrides(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Domain:       "https://default.example.com",
				Organization: "org-default",
				Region:       "center",
			},
			"alt": {
				Domain:       "https://alt.example.com",
				Organization: "org-alt",
				Region:       "center",
			},
		},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("alt", "token-alt"); err != nil {
		t.Fatalf("save token: %v", err)
	}

	runtime, err := (Options{ConfigPath: configPath}).resolveRuntimeProfile(Context{
		Profile:      "alt",
		Domain:       "https://override.example.com",
		Organization: "org-override",
	})
	if err != nil {
		t.Fatalf("expected runtime profile, got: %v", err)
	}
	if runtime.Name != "alt" || runtime.Token != "token-alt" {
		t.Fatalf("unexpected runtime identity: %+v", runtime)
	}
	if runtime.Profile.Domain != "https://override.example.com" || runtime.Profile.Organization != "org-override" {
		t.Fatalf("expected command overrides, got profile %+v", runtime.Profile)
	}
	if runtime.Profile.Region != "center" {
		t.Fatalf("expected persisted region, got %q", runtime.Profile.Region)
	}
}

func TestResolveRuntimeProfileUsesDefaultProfileBeforeCurrent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := config.NewStore(configPath).Save(config.Config{
		Current: "current",
		Profiles: map[string]config.Profile{
			"current": {Domain: "https://current.example.com", Organization: "org-current"},
			"chosen":  {Domain: "https://chosen.example.com", Organization: "org-chosen"},
		},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if err := auth.NewFileTokenStore(filepath.Join(dir, "tokens.yaml")).Save("chosen", "token-chosen"); err != nil {
		t.Fatalf("save token: %v", err)
	}

	runtime, err := (Options{ConfigPath: configPath, DefaultProfile: "chosen"}).resolveRuntimeProfile(Context{})
	if err != nil {
		t.Fatalf("expected runtime profile, got: %v", err)
	}
	if runtime.Name != "chosen" || runtime.Profile.Organization != "org-chosen" || runtime.Token != "token-chosen" {
		t.Fatalf("unexpected runtime profile: %+v", runtime)
	}
}
