package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTopLevelCommandHelpContracts(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "root", args: []string{"--help"}, want: []string{"Available Commands:", "auth", "completion", "config", "issue", "member", "mr", "pipeline", "pr", "project", "repo", "version", "workitem"}},
		{name: "auth", args: []string{"auth", "--help"}, want: []string{"Manage authentication", "status", "login", "logout"}},
		{name: "completion", args: []string{"completion", "--help"}, want: []string{"Generate the autocompletion script", "bash", "zsh", "fish", "powershell"}},
		{name: "config", args: []string{"config", "--help"}, want: []string{"Manage yx configuration", "list", "get", "set", "use"}},
		{name: "issue", args: []string{"issue", "--help"}, want: []string{"Manage Yunxiao work items", "list", "view", "create", "update"}},
		{name: "member", args: []string{"member", "--help"}, want: []string{"Manage Yunxiao organization members", "list", "search", "get"}},
		{name: "mr", args: []string{"mr", "--help"}, want: []string{"Manage Codeup merge requests", "list", "view", "create", "merge"}},
		{name: "pipeline", args: []string{"pipeline", "--help"}, want: []string{"Manage Yunxiao Flow pipelines", "list", "view", "create", "run", "logs"}},
		{name: "pr", args: []string{"pr", "--help"}, want: []string{"Manage Codeup merge requests", "list", "view", "create", "merge"}},
		{name: "project", args: []string{"project", "--help"}, want: []string{"Manage Yunxiao Projex projects", "list", "create"}},
		{name: "repo", args: []string{"repo", "--help"}, want: []string{"Manage Codeup repositories", "list", "current", "view", "create", "clone", "branch", "commit", "file"}},
		{name: "version", args: []string{"version", "--help"}, want: []string{"Show version information"}},
		{name: "workitem", args: []string{"workitem", "--help"}, want: []string{"Manage Yunxiao work items", "list", "view", "create", "update"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml")}), tt.args...)
			if err != nil {
				t.Fatalf("expected %v to succeed, got error: %v stderr=%s", tt.args, err, stderr)
			}
			if !strings.Contains(stdout, "Usage:") {
				t.Fatalf("expected help usage for %v, got:\n%s", tt.args, stdout)
			}
			for _, want := range tt.want {
				if !strings.Contains(stdout, want) {
					t.Fatalf("expected %v help to include %q, got:\n%s", tt.args, want, stdout)
				}
			}
		})
	}
}

func TestHighLevelSubcommandFlagHelpContracts(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "repo current", args: []string{"repo", "current", "--help"}, want: []string{"--remote", "--refresh", "--json"}},
		{name: "repo create", args: []string{"repo", "create", "--help"}, want: []string{"--name", "--path", "--description", "--visibility", "--readme-type", "--dry-run", "--yes"}},
		{name: "repo branch sync", args: []string{"repo", "branch", "sync", "--help"}, want: []string{"--source", "--target", "--dry-run", "--yes"}},
		{name: "repo commit list", args: []string{"repo", "commit", "list", "--help"}, want: []string{"--ref", "--json"}},
		{name: "repo file view", args: []string{"repo", "file", "view", "--help"}, want: []string{"--ref", "--json"}},
		{name: "mr create", args: []string{"mr", "create", "--help"}, want: []string{"--repo", "--source", "--target", "--title", "--dry-run", "--yes"}},
		{name: "mr merge", args: []string{"mr", "merge", "--help"}, want: []string{"--repo", "--dry-run", "--yes"}},
		{name: "project create", args: []string{"project", "create", "--help"}, want: []string{"--name", "--custom-code", "--scope", "--template-id", "--description", "--dry-run", "--yes"}},
		{name: "workitem list", args: []string{"workitem", "list", "--help"}, want: []string{"--project", "--repo"}},
		{name: "member list", args: []string{"member", "list", "--help"}, want: []string{"--status"}},
		{name: "member search", args: []string{"member", "search", "--help"}, want: []string{"--name", "--email", "--status"}},
		{name: "member get", args: []string{"member", "get", "--help"}, want: []string{"--user-id"}},
		{name: "workitem create", args: []string{"workitem", "create", "--help"}, want: []string{"--project", "--type", "--title", "--description", "--description-format", "--assignee", "--dry-run", "--yes"}},
		{name: "workitem update", args: []string{"workitem", "update", "--help"}, want: []string{"--status", "--assignee", "--title", "--description", "--description-format", "--dry-run", "--yes"}},
		{name: "pipeline create", args: []string{"pipeline", "create", "--help"}, want: []string{"--name", "--file", "--dry-run", "--yes"}},
		{name: "pipeline run", args: []string{"pipeline", "run", "--help"}, want: []string{"--branch", "--dry-run", "--yes"}},
		{name: "pipeline run list", args: []string{"pipeline", "run", "list", "--help"}, want: []string{"--branch", "--tag", "--commit", "--page", "--per-page"}},
		{name: "pipeline run logs", args: []string{"pipeline", "run", "logs", "--help"}, want: []string{"--job", "--step-index", "--build-id", "--offset", "--limit"}},
		{name: "pipeline logs", args: []string{"pipeline", "logs", "--help"}, want: []string{"--follow"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml")}), tt.args...)
			if err != nil {
				t.Fatalf("expected %v to succeed, got error: %v stderr=%s", tt.args, err, stderr)
			}
			for _, want := range tt.want {
				if !strings.Contains(stdout, want) {
					t.Fatalf("expected %v help to include %q, got:\n%s", tt.args, want, stdout)
				}
			}
		})
	}
}

func TestCompletionCommandsEmitShellScripts(t *testing.T) {
	tests := []struct {
		shell string
		want  string
	}{
		{shell: "bash", want: "bash completion"},
		{shell: "zsh", want: "compdef"},
		{shell: "fish", want: "complete"},
		{shell: "powershell", want: "Register-ArgumentCompleter"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml")}), "completion", tt.shell)
			if err != nil {
				t.Fatalf("expected completion %s to succeed, got error: %v stderr=%s", tt.shell, err, stderr)
			}
			if !strings.Contains(stdout, tt.want) {
				t.Fatalf("expected completion %s output to include %q, got:\n%s", tt.shell, tt.want, stdout)
			}
		})
	}
}

func TestBuiltBinaryHighLevelSmoke(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	bin := filepath.Join(t.TempDir(), "yx")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/yx")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build yx binary: %v\n%s", err, output)
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "root help", args: nil, want: []string{"Yunxiao command line client", "Available Commands:", "repo", "pipeline"}},
		{name: "version flag", args: []string{"--version"}, want: []string{"yx version"}},
		{name: "repo help", args: []string{"repo", "--help"}, want: []string{"Manage Codeup repositories", "current", "branch"}},
		{name: "completion bash", args: []string{"completion", "bash"}, want: []string{"bash completion"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := exec.Command(bin, tt.args...)
			run.Env = append(os.Environ(), "HOME="+t.TempDir())
			output, err := run.CombinedOutput()
			if err != nil {
				t.Fatalf("expected yx %v to succeed, got error: %v output=%s", tt.args, err, output)
			}
			out := string(output)
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Fatalf("expected yx %v output to include %q, got:\n%s", tt.args, want, out)
				}
			}
		})
	}
}
