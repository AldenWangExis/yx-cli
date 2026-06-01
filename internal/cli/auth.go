package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	yunxiaoPATURL               = "https://account-devops.aliyun.com/settings/personalAccessToken"
	yunxiaoServiceConnectionURL = "https://flow.aliyun.com/setting/service-connection"
)

func newAuthCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}

	cmd.AddCommand(newAuthStatusCommand(opts))
	cmd.AddCommand(newAuthLoginCommand(opts))
	cmd.AddCommand(newAuthLogoutCommand(opts))
	return cmd
}

func newAuthStatusCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := resolveProfile(cmd, opts)
			if err != nil {
				return err
			}
			cfg, err := config.NewStore(opts.ConfigPath).Load()
			if err != nil {
				return err
			}
			profileConfig := cfg.Profiles[profile]
			status, err := opts.authProvider().Status(cmd.Context(), profile)
			if err != nil {
				return err
			}
			account, accountErr := opts.authAccountResolver().ResolveAuthAccount(cmd.Context(), profile, profileConfig)
			return renderAuthStatus(cmd.OutOrStdout(), cfg.Current, profileConfig, status, account, accountErr)
		},
	}
}

func renderAuthStatus(w io.Writer, currentProfile string, profileConfig config.Profile, status auth.Status, account AuthAccount, accountErr error) error {
	host := displayHost(profileConfig.Domain)
	if host == "" {
		host = "yunxiao"
	}

	fmt.Fprintln(w, host)
	if status.HasToken {
		accountName := firstStatusValue(account.Name, account.Username, "unknown")
		fmt.Fprintf(w, "  ✓ Logged in to %s account %s (%s)\n", host, accountName, status.Backend)
	} else {
		fmt.Fprintf(w, "  x Not logged in to %s profile %s (%s)\n", host, status.Profile, status.Backend)
	}

	active := currentProfile == "" || currentProfile == status.Profile
	fmt.Fprintf(w, "  - Active profile: %t\n", active)
	accountStatus := "not authenticated"
	if status.HasToken && accountErr == nil {
		accountStatus = "authenticated"
	} else if status.HasToken {
		accountStatus = "token present, account lookup unavailable"
	}
	fmt.Fprintf(w, "  - Account status: %s\n", accountStatus)
	if account.AccountID != "" {
		fmt.Fprintf(w, "  - Account ID: %s\n", account.AccountID)
	}
	if account.Username != "" && account.Username != account.Name {
		fmt.Fprintf(w, "  - Username: %s\n", account.Username)
	}
	if accountErr != nil {
		fmt.Fprintf(w, "  - Account lookup: unavailable (%v)\n", accountErr)
	}
	if profileConfig.Organization != "" {
		fmt.Fprintf(w, "  - Organization: %s\n", profileConfig.Organization)
	}
	if profileConfig.Region != "" {
		fmt.Fprintf(w, "  - Region: %s\n", profileConfig.Region)
	}
	token := "missing"
	if status.TokenMask != "" {
		token = status.TokenMask
	} else if status.HasToken {
		token = "present"
	}
	fmt.Fprintf(w, "  - Token: %s\n", token)
	if !status.HasToken {
		fmt.Fprintf(w, "  - Personal access token: %s\n", yunxiaoPATURL)
		fmt.Fprintf(w, "  - Service connections: %s\n", yunxiaoServiceConnectionURL)
	}
	if len(profileConfig.ServiceConnections) > 0 {
		fmt.Fprintln(w, "  - Service connections:")
		keys := make([]string, 0, len(profileConfig.ServiceConnections))
		for key := range profileConfig.ServiceConnections {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(w, "    %s: %s\n", key, maskStatusSecret(profileConfig.ServiceConnections[key]))
		}
	}
	return nil
}

func displayHost(domain string) string {
	value := strings.TrimSpace(domain)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed.Host != "" {
		return parsed.Host
	}
	return strings.TrimRight(strings.TrimPrefix(strings.TrimPrefix(value, "https://"), "http://"), "/")
}

func firstStatusValue(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func maskStatusSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	const prefixLen = 3
	const suffixLen = 4
	if len(value) <= prefixLen+suffixLen {
		return strings.Repeat("*", len(value))
	}
	return value[:prefixLen] + strings.Repeat("*", len(value)-prefixLen-suffixLen) + value[len(value)-suffixLen:]
}

func newAuthLoginCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login with a Yunxiao personal access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := resolveProfile(cmd, opts)
			if err != nil {
				return err
			}
			reader := bufio.NewReader(cmd.InOrStdin())
			fmt.Fprintf(cmd.OutOrStdout(), "Yunxiao personal access token (%s): ", yunxiaoPATURL)
			token, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("read token: %w", err)
			}
			token = strings.TrimSpace(token)
			if token == "" {
				return fmt.Errorf("token is required")
			}
			status, err := opts.authProvider().Login(cmd.Context(), profile, token)
			if err != nil {
				return err
			}
			serviceConnection, err := readOptionalLine(reader, cmd.OutOrStdout(), fmt.Sprintf("Codeup service connection ID (optional, %s): ", yunxiaoServiceConnectionURL))
			if err != nil {
				return err
			}
			if serviceConnection != "" {
				if err := saveCodeupServiceConnection(opts.ConfigPath, profile, serviceConnection); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "stored Codeup service connection for profile %s\n", status.Profile)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "logged in profile %s using %s token store\n", status.Profile, status.Backend)
			return nil
		},
	}
}

func readOptionalLine(reader *bufio.Reader, w io.Writer, prompt string) (string, error) {
	fmt.Fprint(w, prompt)
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read optional value: %w", err)
	}
	return strings.TrimSpace(value), nil
}

func saveCodeupServiceConnection(configPath, profileName, id string) error {
	store := config.NewStore(configPath)
	cfg, err := store.Load()
	if err != nil {
		return err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]config.Profile{}
	}
	profile := cfg.Profiles[profileName]
	if profile.ServiceConnections == nil {
		profile.ServiceConnections = map[string]string{}
	}
	profile.ServiceConnections["codeup"] = id
	cfg.Profiles[profileName] = profile
	if cfg.Current == "" {
		cfg.Current = profileName
	}
	return store.Save(cfg)
}

func newAuthLogoutCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := resolveProfile(cmd, opts)
			if err != nil {
				return err
			}
			if err := opts.authProvider().Logout(cmd.Context(), profile); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "logged out profile %s\n", profile)
			return nil
		},
	}
}

func resolveProfile(cmd *cobra.Command, opts Options) (string, error) {
	if profile := ContextFromCommand(cmd).Profile; profile != "" {
		return profile, nil
	}
	if opts.DefaultProfile != "" {
		return opts.DefaultProfile, nil
	}
	cfg, err := config.NewStore(opts.ConfigPath).Load()
	if err != nil {
		return "", err
	}
	if cfg.Current != "" {
		return cfg.Current, nil
	}
	return "default", nil
}

func defaultTokenPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "tokens.yaml")
}
