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

func TestRootUpdateCheckWritesHintToStderr(t *testing.T) {
	oldVersion := Version
	Version = "v1.0.0"
	t.Cleanup(func() { Version = oldVersion })

	checker := &fakeUpdateChecker{notice: UpdateNotice{Available: true, Current: "v1.0.0", Latest: "v1.1.0", Command: "yx self update"}}
	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{UpdateChecker: checker}), "version")
	if err != nil {
		t.Fatalf("version: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "yx version v1.0.0") {
		t.Fatalf("expected version on stdout, got %s", stdout)
	}
	for _, want := range []string{"newer yx is available", "v1.0.0 -> v1.1.0", "yx self update"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("expected update hint %q in stderr, got:\n%s", want, stderr)
		}
	}
	if checker.calls != 1 {
		t.Fatalf("expected one update check, got %d", checker.calls)
	}
}

func TestRootUpdateCheckSkipsJSONOutput(t *testing.T) {
	checker := &fakeUpdateChecker{notice: UpdateNotice{Available: true, Current: "v1.0.0", Latest: "v1.1.0"}}
	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{UpdateChecker: checker}), "--json", "config", "list")
	if err != nil {
		t.Fatalf("config list: %v stderr=%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected no update stderr for json output, got %s", stderr)
	}
	if checker.calls != 0 {
		t.Fatalf("expected no update check for json output, got %d", checker.calls)
	}
}

func TestRootUpdateCheckSkipsWhenDisabled(t *testing.T) {
	t.Setenv(updateCheckOptOutEnv, "1")
	checker := &fakeUpdateChecker{notice: UpdateNotice{Available: true, Current: "v1.0.0", Latest: "v1.1.0"}}
	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{UpdateChecker: checker}), "version")
	if err != nil {
		t.Fatalf("version: %v stderr=%s", err, stderr)
	}
	if stderr != "" || checker.calls != 0 {
		t.Fatalf("expected disabled update check, stderr=%q calls=%d", stderr, checker.calls)
	}
}

type fakeUpdateChecker struct {
	notice UpdateNotice
	calls  int
}

func (c *fakeUpdateChecker) Check(ctx UpdateCheckContext) (UpdateNotice, error) {
	c.calls++
	return c.notice, nil
}
