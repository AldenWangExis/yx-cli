package gitx

import (
	"context"
	"testing"
)

func TestRunnerRejectsTokenInCloneURL(t *testing.T) {
	runner := NewRunner()

	err := runner.Clone(context.Background(), "https://oauth2:secret-token@example.com/org/repo.git", "")
	if err == nil {
		t.Fatal("expected clone URL with token-like credentials to fail")
	}
}
