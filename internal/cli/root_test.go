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
