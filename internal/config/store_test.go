package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingConfigReturnsEmptyConfig(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "config.yaml"))

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("expected missing config to load, got error: %v", err)
	}
	if cfg.Current != "" {
		t.Fatalf("expected empty current profile, got %q", cfg.Current)
	}
	if cfg.Profiles == nil {
		t.Fatal("expected profiles map to be initialized")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	store := NewStore(path)
	cfg := Config{
		Current: "default",
		Profiles: map[string]Profile{
			"default": {
				Domain:       "https://devops.aliyun.com",
				Organization: "org-1",
				Region:       "center",
				Output:       "table",
				Safety: Safety{
					ConfirmWrites: true,
				},
				ServiceConnections: map[string]string{
					"codeup": "ijhr7pdz5567r9p6",
				},
				RepoProjectMap: map[string]string{
					"repo-a": "project-a",
				},
			},
		},
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("expected save to succeed, got error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("expected load to succeed, got error: %v", err)
	}
	if loaded.Current != "default" {
		t.Fatalf("expected current profile default, got %q", loaded.Current)
	}
	if loaded.Profiles["default"].Domain != "https://devops.aliyun.com" {
		t.Fatalf("expected domain to round-trip, got %q", loaded.Profiles["default"].Domain)
	}
	if !loaded.Profiles["default"].Safety.ConfirmWrites {
		t.Fatal("expected safety.confirmWrites to round-trip")
	}
	if loaded.Profiles["default"].ServiceConnections["codeup"] != "ijhr7pdz5567r9p6" {
		t.Fatalf("expected service connection to round-trip")
	}
	if loaded.Profiles["default"].RepoProjectMap["repo-a"] != "project-a" {
		t.Fatalf("expected repo project mapping to round-trip")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected saved file to exist, got error: %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("expected config permissions to reject group/world access, got %v", info.Mode().Perm())
	}
}

func TestLoadRejectsUnknownKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("current: default\nunknown: true\n"), 0o600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := NewStore(path).Load()
	if err == nil {
		t.Fatal("expected unknown top-level key to fail")
	}
}

func TestAtomicSaveKeepsOldConfigOnFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	store := NewStore(path)

	if err := store.Save(Config{
		Current:  "old",
		Profiles: map[string]Profile{"old": {Domain: "https://old.example.com"}},
	}); err != nil {
		t.Fatalf("expected initial save to succeed, got error: %v", err)
	}

	store.beforeRename = func() error {
		return errors.New("injected failure")
	}
	err := store.Save(Config{
		Current:  "new",
		Profiles: map[string]Profile{"new": {Domain: "https://new.example.com"}},
	})
	if err == nil {
		t.Fatal("expected injected failure")
	}

	loaded, err := NewStore(path).Load()
	if err != nil {
		t.Fatalf("expected old config to remain loadable, got error: %v", err)
	}
	if loaded.Current != "old" {
		t.Fatalf("expected old config to remain active, got %q", loaded.Current)
	}
	if _, ok := loaded.Profiles["new"]; ok {
		t.Fatal("expected failed save not to publish new profile")
	}
}
