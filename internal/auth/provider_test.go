package auth

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestPATProviderStatusMasksToken(t *testing.T) {
	store := NewFileTokenStore(filepath.Join(t.TempDir(), "tokens.yaml"))
	if err := store.Save("default", "pt-example-token-body-for-mask-test-b75c"); err != nil {
		t.Fatalf("save token: %v", err)
	}

	status, err := NewPATProvider(store).Status(context.Background(), "default")
	if err != nil {
		t.Fatalf("status: %v", err)
	}

	if !status.HasToken {
		t.Fatal("expected token to be present")
	}
	if status.TokenMask == "" {
		t.Fatal("expected masked token")
	}
	if strings.Contains(status.TokenMask, "example-token-body") {
		t.Fatalf("token mask leaked token body: %q", status.TokenMask)
	}
	if !strings.HasPrefix(status.TokenMask, "pt-") || !strings.HasSuffix(status.TokenMask, "b75c") {
		t.Fatalf("expected token mask to keep prefix and suffix, got %q", status.TokenMask)
	}
}
