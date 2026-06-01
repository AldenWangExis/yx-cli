package gitx

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Clone(ctx context.Context, cloneURL, destination string) error {
	if hasTokenLikeCredentials(cloneURL) {
		return fmt.Errorf("clone URL contains credentials; use SSH or a git credential helper")
	}
	args := []string{"clone", cloneURL}
	if destination != "" {
		args = append(args, destination)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s", redactCloneURL(string(output), cloneURL))
	}
	return nil
}

func hasTokenLikeCredentials(cloneURL string) bool {
	parsed, err := url.Parse(cloneURL)
	if err != nil || parsed.User == nil {
		return false
	}
	username := strings.ToLower(parsed.User.Username())
	_, hasPassword := parsed.User.Password()
	return hasPassword || strings.Contains(username, "token") || strings.Contains(username, "oauth")
}

func redactCloneURL(value, cloneURL string) string {
	if cloneURL == "" {
		return value
	}
	return strings.ReplaceAll(value, cloneURL, "[REDACTED_CLONE_URL]")
}
