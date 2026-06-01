package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected help to succeed, got error: %v", err)
	}

	out := stdout.String()
	for _, want := range []string{
		"Yunxiao command line client",
		"Usage:",
		"Available Commands:",
		"repo",
		"mr",
		"pipeline",
		"Flags:",
		"yx repo list",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected help to include %q, got:\n%s", want, out)
		}
	}
	if stderr.String() != "" {
		t.Fatalf("expected no stderr, got:\n%s", stderr.String())
	}
}

func TestRootVersionFlagAndCommand(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	Version, Commit, Date = "v9.9.9", "abcdef1", "2026-06-01T00:00:00Z"
	t.Cleanup(func() {
		Version, Commit, Date = oldVersion, oldCommit, oldDate
	})

	for _, args := range [][]string{{"--version"}, {"version"}} {
		stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{}), args...)
		if err != nil {
			t.Fatalf("expected %v to succeed, got error: %v stderr=%s", args, err, stderr)
		}
		for _, want := range []string{"yx version v9.9.9", "abcdef1", "2026-06-01T00:00:00Z"} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("expected version output to include %q, got:\n%s", want, stdout)
			}
		}
	}
}
