package cli

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/spf13/cobra"
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
			status, err := opts.authProvider().Status(cmd.Context(), profile)
			if err != nil {
				return err
			}
			tokenState := "missing"
			if status.HasToken {
				tokenState = "present"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "profile: %s\nbackend: %s\ntoken: %s\n", status.Profile, status.Backend, tokenState)
			return nil
		},
	}
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
			token, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
			if err != nil {
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
			fmt.Fprintf(cmd.OutOrStdout(), "logged in profile %s using %s token store\n", status.Profile, status.Backend)
			return nil
		},
	}
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
