package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/platform"
	"github.com/spf13/cobra"
)

type Options struct {
	ConfigPath          string
	AuthProvider        auth.Provider
	DefaultProfile      string
	RepoUseCase         RepositoryUseCase
	RepoCurrentResolver RepositoryCurrentResolver
	MergeRequestUseCase MergeRequestUseCase
	WorkitemUseCase     WorkitemUseCase
	PipelineUseCase     PipelineUseCase
	AuthAccountResolver AuthAccountResolver
}

type AuthAccount struct {
	Name      string
	Username  string
	AccountID string
	Email     string
}

type AuthAccountResolver interface {
	ResolveAuthAccount(ctx context.Context, profileName string, profile config.Profile) (AuthAccount, error)
}

func defaultOptions() Options {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return Options{ConfigPath: filepath.Join(home, ".config", "yx", "config.yaml")}
}

func (o Options) authProvider() auth.Provider {
	if o.AuthProvider != nil {
		return o.AuthProvider
	}
	return auth.NewPATProvider(auth.NewFileTokenStore(defaultTokenPath(o.ConfigPath)))
}

func (o Options) authAccountResolver() AuthAccountResolver {
	if o.AuthAccountResolver != nil {
		return o.AuthAccountResolver
	}
	return fileAuthAccountResolver{configPath: o.ConfigPath}
}

type fileAuthAccountResolver struct {
	configPath string
}

func (r fileAuthAccountResolver) ResolveAuthAccount(ctx context.Context, profileName string, profile config.Profile) (AuthAccount, error) {
	token, ok, err := auth.NewFileTokenStore(defaultTokenPath(r.configPath)).Load(profileName)
	if err != nil || !ok || token == "" || profile.Domain == "" {
		return AuthAccount{}, err
	}
	user, err := platform.NewAdapter(yunxiao.ClientConfig{
		BaseURL:        profile.Domain,
		Token:          token,
		OrganizationID: profile.Organization,
		Region:         profile.Region,
	}).CurrentUser(ctx)
	if err != nil {
		return AuthAccount{}, err
	}
	return AuthAccount{Name: user.Name, Username: user.Username, AccountID: user.AccountID, Email: user.Email}, nil
}

func newConfigCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage yx configuration",
	}

	cmd.AddCommand(newConfigListCommand(opts))
	cmd.AddCommand(newConfigGetCommand(opts))
	cmd.AddCommand(newConfigSetCommand(opts))
	cmd.AddCommand(newConfigUseCommand(opts))

	return cmd
}

func newConfigListCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.NewStore(opts.ConfigPath).Load()
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(cfg)
			}
			return renderer.WriteTable([]string{"CURRENT", "PROFILES"}, [][]string{{cfg.Current, strconv.Itoa(len(cfg.Profiles))}})
		},
	}
}

func newConfigGetCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.NewStore(opts.ConfigPath).Load()
			if err != nil {
				return err
			}
			value, ok := getValue(cfg, args[0])
			if !ok {
				return fmt.Errorf("unknown config key %q", args[0])
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}
}

func newConfigSetCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := config.NewStore(opts.ConfigPath)
			cfg, err := store.Load()
			if err != nil {
				return err
			}
			if err := setValue(&cfg, args[0], args[1]); err != nil {
				return err
			}
			return store.Save(cfg)
		},
	}
}

func newConfigUseCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Set the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := config.NewStore(opts.ConfigPath)
			cfg, err := store.Load()
			if err != nil {
				return err
			}
			if _, ok := cfg.Profiles[args[0]]; !ok {
				return fmt.Errorf("profile %q does not exist", args[0])
			}
			cfg.Current = args[0]
			return store.Save(cfg)
		},
	}
}

func getValue(cfg config.Config, key string) (string, bool) {
	switch key {
	case "current":
		return cfg.Current, true
	}

	profileName, field, ok := splitProfileKey(key)
	if !ok {
		return "", false
	}
	profile, ok := cfg.Profiles[profileName]
	if !ok {
		return "", false
	}

	switch field {
	case "domain":
		return profile.Domain, true
	case "organization":
		return profile.Organization, true
	case "region":
		return profile.Region, true
	case "output":
		return profile.Output, true
	case "safety.confirmWrites":
		return strconv.FormatBool(profile.Safety.ConfirmWrites), true
	default:
		const serviceConnectionPrefix = "serviceConnections."
		if len(field) > len(serviceConnectionPrefix) && field[:len(serviceConnectionPrefix)] == serviceConnectionPrefix {
			value, ok := profile.ServiceConnections[field[len(serviceConnectionPrefix):]]
			return value, ok
		}
		const prefix = "repoProjectMap."
		if len(field) > len(prefix) && field[:len(prefix)] == prefix {
			value, ok := profile.RepoProjectMap[field[len(prefix):]]
			return value, ok
		}
		return "", false
	}
}

func setValue(cfg *config.Config, key, value string) error {
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]config.Profile{}
	}
	if key == "current" {
		cfg.Current = value
		return nil
	}

	profileName, field, ok := splitProfileKey(key)
	if !ok {
		return fmt.Errorf("unknown config key %q", key)
	}
	profile := cfg.Profiles[profileName]
	if profile.ServiceConnections == nil {
		profile.ServiceConnections = map[string]string{}
	}
	if profile.RepoProjectMap == nil {
		profile.RepoProjectMap = map[string]string{}
	}

	switch field {
	case "domain":
		profile.Domain = value
	case "organization":
		profile.Organization = value
	case "region":
		profile.Region = value
	case "output":
		profile.Output = value
	case "safety.confirmWrites":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse safety.confirmWrites: %w", err)
		}
		profile.Safety.ConfirmWrites = parsed
	default:
		const serviceConnectionPrefix = "serviceConnections."
		if len(field) > len(serviceConnectionPrefix) && field[:len(serviceConnectionPrefix)] == serviceConnectionPrefix {
			profile.ServiceConnections[field[len(serviceConnectionPrefix):]] = value
			break
		}
		const prefix = "repoProjectMap."
		if len(field) <= len(prefix) || field[:len(prefix)] != prefix {
			return fmt.Errorf("unknown config key %q", key)
		}
		profile.RepoProjectMap[field[len(prefix):]] = value
	}

	cfg.Profiles[profileName] = profile
	return nil
}

func splitProfileKey(key string) (string, string, bool) {
	const prefix = "profiles."
	if len(key) <= len(prefix) || key[:len(prefix)] != prefix {
		return "", "", false
	}
	rest := key[len(prefix):]
	for i, ch := range rest {
		if ch == '.' {
			if i == 0 || i == len(rest)-1 {
				return "", "", false
			}
			return rest[:i], rest[i+1:], true
		}
	}
	return "", "", false
}
