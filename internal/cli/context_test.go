package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestGlobalFlagsPopulateCommandContext(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var got Context

	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.AddCommand(&cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, args []string) error {
			got = ContextFromCommand(cmd)
			return nil
		},
	})
	cmd.SetArgs([]string{
		"--profile", "dev",
		"--org", "org-1",
		"--domain", "https://devops.example.com",
		"--json",
		"--verbose",
		"probe",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected command to succeed, got error: %v", err)
	}

	if got.Profile != "dev" {
		t.Fatalf("expected profile dev, got %q", got.Profile)
	}
	if got.Organization != "org-1" {
		t.Fatalf("expected organization org-1, got %q", got.Organization)
	}
	if got.Domain != "https://devops.example.com" {
		t.Fatalf("expected domain override, got %q", got.Domain)
	}
	if !got.JSON {
		t.Fatal("expected JSON output flag to be true")
	}
	if !got.Verbose {
		t.Fatal("expected verbose flag to be true")
	}
	if stderr.String() != "" {
		t.Fatalf("expected no stderr, got:\n%s", stderr.String())
	}
}

func TestFlagParseErrorsUseStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--does-not-exist"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected unknown flag to fail")
	}
	if stdout.String() != "" {
		t.Fatalf("expected no stdout, got:\n%s", stdout.String())
	}
	if stderr.String() == "" {
		t.Fatal("expected flag parse error on stderr")
	}
}
