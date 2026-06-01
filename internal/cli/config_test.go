package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/spf13/cobra"
)

func TestConfigCommandsSetGetUseAndList(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"config", "set", "profiles.default.domain", "https://devops.aliyun.com")
	if err != nil {
		t.Fatalf("expected config set to succeed, got error: %v stderr=%s", err, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected config set to write no stdout, got:\n%s", stdout)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"config", "get", "profiles.default.domain")
	if err != nil {
		t.Fatalf("expected config get to succeed, got error: %v stderr=%s", err, stderr)
	}
	if strings.TrimSpace(stdout) != "https://devops.aliyun.com" {
		t.Fatalf("expected config get to print domain, got:\n%s", stdout)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"config", "set", "profiles.default.serviceConnections.codeup", "ijhr7pdz5567r9p6")
	if err != nil {
		t.Fatalf("expected service connection config set to succeed, got error: %v stderr=%s", err, stderr)
	}
	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"config", "get", "profiles.default.serviceConnections.codeup")
	if err != nil {
		t.Fatalf("expected service connection config get to succeed, got error: %v stderr=%s", err, stderr)
	}
	if strings.TrimSpace(stdout) != "ijhr7pdz5567r9p6" {
		t.Fatalf("expected config get to print service connection, got:\n%s", stdout)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"config", "use", "default")
	if err != nil {
		t.Fatalf("expected config use to succeed, got error: %v stderr=%s", err, stderr)
	}

	loaded, err := config.NewStore(configPath).Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}
	if loaded.Current != "default" {
		t.Fatalf("expected current profile default, got %q", loaded.Current)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: configPath}),
		"--json", "config", "list")
	if err != nil {
		t.Fatalf("expected config list --json to succeed, got error: %v stderr=%s", err, stderr)
	}
	var listed config.Config
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatalf("expected config list --json to emit JSON, got error: %v output=%s", err, stdout)
	}
	if listed.Profiles["default"].Domain != "https://devops.aliyun.com" {
		t.Fatalf("expected JSON config to include profile domain")
	}
	if listed.Profiles["default"].ServiceConnections["codeup"] != "ijhr7pdz5567r9p6" {
		t.Fatalf("expected JSON config to include service connection")
	}
}

func TestConfigCommandMissingArgsFail(t *testing.T) {
	_, stderr, err := executeCommand(t, NewRootCommandWithOptions(Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml")}),
		"config", "get")
	if err == nil {
		t.Fatal("expected config get without key to fail")
	}
	if stderr == "" {
		t.Fatal("expected missing args error on stderr")
	}
}

func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}
