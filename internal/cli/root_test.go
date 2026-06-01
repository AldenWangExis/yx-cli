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
	if !strings.Contains(out, "yx") {
		t.Fatalf("expected help to include command name, got:\n%s", out)
	}
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("expected help to include usage, got:\n%s", out)
	}
	if stderr.String() != "" {
		t.Fatalf("expected no stderr, got:\n%s", stderr.String())
	}
}
