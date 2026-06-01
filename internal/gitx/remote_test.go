package gitx

import "testing"

func TestParseCodeupRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		wantOrg  string
		wantPath string
	}{
		{
			name:     "scp ssh url",
			rawURL:   "git@codeup.aliyun.com:68086322e3a71588779435e0/yx-cli.git",
			wantOrg:  "68086322e3a71588779435e0",
			wantPath: "68086322e3a71588779435e0/yx-cli",
		},
		{
			name:     "ssh url",
			rawURL:   "ssh://git@codeup.aliyun.com/68086322e3a71588779435e0/group/yx-cli.git",
			wantOrg:  "68086322e3a71588779435e0",
			wantPath: "68086322e3a71588779435e0/group/yx-cli",
		},
		{
			name:     "https url",
			rawURL:   "https://codeup.aliyun.com/68086322e3a71588779435e0/yx-cli.git",
			wantOrg:  "68086322e3a71588779435e0",
			wantPath: "68086322e3a71588779435e0/yx-cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remote, err := ParseCodeupRemoteURL("origin", tt.rawURL)
			if err != nil {
				t.Fatalf("expected parse to succeed, got: %v", err)
			}
			if remote.Organization != tt.wantOrg {
				t.Fatalf("expected organization %q, got %q", tt.wantOrg, remote.Organization)
			}
			if remote.PathWithNamespace != tt.wantPath {
				t.Fatalf("expected path %q, got %q", tt.wantPath, remote.PathWithNamespace)
			}
			if remote.RepositoryName != "yx-cli" {
				t.Fatalf("expected repository name yx-cli, got %q", remote.RepositoryName)
			}
		})
	}
}

func TestParseCodeupRemoteURLRejectsUnsupportedOrUnsafeURLs(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
	}{
		{name: "github remote", rawURL: "git@github.com:AldenWangExis/yx-cli.git"},
		{name: "missing repository path", rawURL: "git@codeup.aliyun.com:68086322e3a71588779435e0.git"},
		{name: "https credentials", rawURL: "https://oauth2:secret-token@codeup.aliyun.com/68086322e3a71588779435e0/yx-cli.git"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseCodeupRemoteURL("origin", tt.rawURL); err == nil {
				t.Fatalf("expected %q to be rejected", tt.rawURL)
			}
		})
	}
}

func TestParseGitRemoteVerboseOutputDeduplicatesFetchAndPush(t *testing.T) {
	remotes := ParseRemoteVerboseOutput(`origin	git@codeup.aliyun.com:org/repo.git (fetch)
origin	git@codeup.aliyun.com:org/repo.git (push)
github	git@github.com:org/repo.git (fetch)
`)

	if len(remotes) != 2 {
		t.Fatalf("expected two remotes, got %+v", remotes)
	}
	if remotes[0].Name != "origin" || remotes[0].URL != "git@codeup.aliyun.com:org/repo.git" {
		t.Fatalf("unexpected first remote: %+v", remotes[0])
	}
	if remotes[1].Name != "github" {
		t.Fatalf("unexpected second remote: %+v", remotes[1])
	}
}
