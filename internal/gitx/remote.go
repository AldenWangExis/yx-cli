package gitx

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

type Remote struct {
	Name string
	URL  string
}

type CodeupRemote struct {
	RemoteName        string
	RemoteURL         string
	Host              string
	Organization      string
	PathWithNamespace string
	RepositoryName    string
}

func (r *Runner) ListRemotes(ctx context.Context, workDir string) ([]Remote, error) {
	args := []string{"remote", "-v"}
	if workDir != "" {
		args = append([]string{"-C", workDir}, args...)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git remote -v failed: %s", strings.TrimSpace(string(output)))
	}
	return ParseRemoteVerboseOutput(string(output)), nil
}

func ParseRemoteVerboseOutput(output string) []Remote {
	remotes := []Remote{}
	seen := map[string]bool{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		if seen[name] {
			continue
		}
		seen[name] = true
		remotes = append(remotes, Remote{Name: name, URL: fields[1]})
	}
	return remotes
}

func ParseCodeupRemoteURL(remoteName, rawURL string) (CodeupRemote, error) {
	host, path, err := splitGitRemoteURL(rawURL)
	if err != nil {
		return CodeupRemote{}, err
	}
	if strings.ToLower(host) != "codeup.aliyun.com" {
		return CodeupRemote{}, fmt.Errorf("remote %q is not a Codeup remote", remoteName)
	}
	path = strings.Trim(strings.TrimSuffix(path, ".git"), "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" || parts[len(parts)-1] == "" {
		return CodeupRemote{}, fmt.Errorf("remote %q does not include organization and repository path", remoteName)
	}
	return CodeupRemote{
		RemoteName:        remoteName,
		RemoteURL:         rawURL,
		Host:              strings.ToLower(host),
		Organization:      parts[0],
		PathWithNamespace: strings.Join(parts, "/"),
		RepositoryName:    parts[len(parts)-1],
	}, nil
}

func splitGitRemoteURL(rawURL string) (host string, path string, err error) {
	if hasScheme(rawURL) {
		parsed, parseErr := url.Parse(rawURL)
		if parseErr != nil {
			return "", "", fmt.Errorf("parse remote URL: %w", parseErr)
		}
		if parsed.User != nil && (parsed.Scheme == "http" || parsed.Scheme == "https" || hasTokenLikeCredentials(rawURL)) {
			return "", "", fmt.Errorf("remote URL contains credentials; use SSH or a git credential helper")
		}
		unescapedPath, pathErr := url.PathUnescape(parsed.Path)
		if pathErr != nil {
			return "", "", fmt.Errorf("parse remote path: %w", pathErr)
		}
		return parsed.Hostname(), unescapedPath, nil
	}

	at := strings.Index(rawURL, "@")
	colon := strings.Index(rawURL, ":")
	if at <= 0 || colon <= at+1 {
		return "", "", fmt.Errorf("unsupported git remote URL")
	}
	user := strings.ToLower(rawURL[:at])
	if strings.Contains(user, "token") || strings.Contains(user, "oauth") {
		return "", "", fmt.Errorf("remote URL contains credentials; use SSH or a git credential helper")
	}
	return rawURL[at+1 : colon], rawURL[colon+1:], nil
}

func hasScheme(rawURL string) bool {
	return strings.Contains(rawURL, "://")
}
