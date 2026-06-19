package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{current: "v1.2.1", latest: "v1.4.0", want: true},
		{current: "v1.4.0", latest: "v1.4.0", want: false},
		{current: "dev", latest: "v1.4.0", want: false},
		{current: "v1.10.0", latest: "v1.9.9", want: false},
	}
	for _, tt := range tests {
		if got := isNewerVersion(tt.current, tt.latest); got != tt.want {
			t.Fatalf("isNewerVersion(%q, %q)=%v want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestGitHubUpdateCheckerUsesCache(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"tag_name":"v1.4.0"}`))
	}))
	defer server.Close()

	checker := NewGitHubUpdateChecker(UpdateCheckerConfig{
		LatestURL:     server.URL,
		CachePath:     filepath.Join(t.TempDir(), "update.json"),
		Current:       "v1.0.0",
		Now:           func() time.Time { return time.Unix(1000, 0).UTC() },
		CacheDuration: 24 * time.Hour,
	})

	notice, err := checker.Check(UpdateCheckContext{Context: context.Background()})
	if err != nil {
		t.Fatalf("first check: %v", err)
	}
	if !notice.Available || notice.Latest != "v1.4.0" || calls != 1 {
		t.Fatalf("unexpected first notice=%+v calls=%d", notice, calls)
	}

	notice, err = checker.Check(UpdateCheckContext{Context: context.Background()})
	if err != nil {
		t.Fatalf("cached check: %v", err)
	}
	if !notice.Available || calls != 1 {
		t.Fatalf("expected cache hit, notice=%+v calls=%d", notice, calls)
	}
}

func TestGitHubUpdateCheckerUsesNPMUpdateCommandForNPMInstall(t *testing.T) {
	t.Setenv("YX_INSTALL_CHANNEL", "npm")
	t.Setenv("YX_NPM_PACKAGE", "@aldenwangexis/yx-cli")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.4.0"}`))
	}))
	defer server.Close()

	checker := NewGitHubUpdateChecker(UpdateCheckerConfig{
		LatestURL: server.URL,
		CachePath: filepath.Join(t.TempDir(), "update.json"),
		Current:   "v1.0.0",
	})

	notice, err := checker.Check(UpdateCheckContext{Context: context.Background()})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if notice.Command != "npm update -g @aldenwangexis/yx-cli" {
		t.Fatalf("unexpected update command: %q", notice.Command)
	}
}
