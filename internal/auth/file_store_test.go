package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileTokenStoreSaveLoadDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.yaml")
	store := NewFileTokenStore(path)

	if err := store.Save("default", "secret-token"); err != nil {
		t.Fatalf("expected save to succeed, got: %v", err)
	}

	token, ok, err := store.Load("default")
	if err != nil {
		t.Fatalf("expected load to succeed, got: %v", err)
	}
	if !ok {
		t.Fatal("expected token to exist")
	}
	if token != "secret-token" {
		t.Fatalf("expected token to round-trip")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected token file to exist, got: %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("expected token file to reject group/world access, got %v", info.Mode().Perm())
	}

	if err := store.Delete("default"); err != nil {
		t.Fatalf("expected delete to succeed, got: %v", err)
	}
	_, ok, err = store.Load("default")
	if err != nil {
		t.Fatalf("expected load after delete to succeed, got: %v", err)
	}
	if ok {
		t.Fatal("expected token to be deleted")
	}
}

func TestFileTokenStoreAtomicSaveKeepsOldTokenOnFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.yaml")
	store := NewFileTokenStore(path)

	if err := store.Save("default", "old-token"); err != nil {
		t.Fatalf("expected initial save to succeed, got: %v", err)
	}

	store.beforeRename = func() error {
		return errors.New("injected failure")
	}
	err := store.Save("default", "new-token")
	if err == nil {
		t.Fatal("expected injected failure")
	}
	if strings.Contains(err.Error(), "new-token") {
		t.Fatal("expected error not to expose token")
	}

	token, ok, err := NewFileTokenStore(path).Load("default")
	if err != nil {
		t.Fatalf("expected old token to remain loadable, got: %v", err)
	}
	if !ok || token != "old-token" {
		t.Fatalf("expected old token to remain, ok=%v token=%q", ok, token)
	}
}
