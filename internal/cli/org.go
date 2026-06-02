package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/platform"
	"github.com/spf13/cobra"
)

const defaultYunxiaoDomain = "https://devops.aliyun.com"

type OrganizationUseCase interface {
	ListOrganizations(ctx context.Context) ([]app.Organization, error)
}

func newOrganizationCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "org",
		Aliases: []string{"organization"},
		Short:   "Manage Yunxiao organizations",
		Example: "  yx org list\n  yx --json org list\n  yx org use <organization-id>",
	}
	cmd.AddCommand(newOrganizationListCommand(opts))
	cmd.AddCommand(newOrganizationUseCommand(opts))
	return cmd
}

func newOrganizationListCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List organizations available to the active token",
		RunE: func(cmd *cobra.Command, args []string) error {
			usecase, err := opts.resolveOrganizationUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			organizations, err := usecase.ListOrganizations(cmd.Context())
			if err != nil {
				return err
			}
			return renderOrganizations(cmd.OutOrStdout(), ContextFromCommand(cmd).JSON, organizations)
		},
	}
}

func newOrganizationUseCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "use <organization-id>",
		Short: "Set the active profile organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := resolveProfile(cmd, opts)
			if err != nil {
				return err
			}
			if err := saveProfileOrganization(opts.ConfigPath, profile, app.Organization{ID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "set profile %s organization to %s\n", profile, args[0])
			return nil
		},
	}
}

func renderOrganizations(w interface{ Write([]byte) (int, error) }, jsonOutput bool, organizations []app.Organization) error {
	renderer := output.NewRenderer(w)
	if jsonOutput {
		return renderer.WriteJSON(organizations)
	}
	rows := make([][]string, 0, len(organizations))
	for i, organization := range organizations {
		rows = append(rows, []string{strconv.Itoa(i + 1), organization.ID, organization.Name})
	}
	return renderer.WriteTable([]string{"#", "ID", "NAME"}, rows)
}

func (o Options) resolveOrganizationUseCase(ctx Context) (OrganizationUseCase, error) {
	if o.OrganizationUseCase != nil {
		return o.OrganizationUseCase, nil
	}
	_, profile, token, err := o.resolveOrganizationProfile(ctx)
	if err != nil {
		return nil, err
	}
	return app.NewOrganizationUseCase(platform.NewAdapter(yunxiao.ClientConfig{
		BaseURL:        profile.Domain,
		Token:          token,
		OrganizationID: profile.Organization,
		Region:         profile.Region,
	})), nil
}

func (o Options) resolveOrganizationProfile(ctx Context) (string, config.Profile, string, error) {
	store := config.NewStore(o.ConfigPath)
	cfg, err := store.Load()
	if err != nil {
		return "", config.Profile{}, "", err
	}
	profileName := ctx.Profile
	if profileName == "" {
		profileName = o.DefaultProfile
	}
	if profileName == "" {
		profileName = cfg.Current
	}
	if profileName == "" {
		profileName = "default"
	}
	profile := cfg.Profiles[profileName]
	if ctx.Domain != "" {
		profile.Domain = ctx.Domain
	}
	if ctx.Organization != "" {
		profile.Organization = ctx.Organization
	}
	if profile.Domain == "" {
		profile.Domain = defaultYunxiaoDomain
	}
	token, ok, err := auth.NewFileTokenStore(defaultTokenPath(o.ConfigPath)).Load(profileName)
	if err != nil {
		return "", config.Profile{}, "", err
	}
	if !ok || token == "" {
		return "", config.Profile{}, "", fmt.Errorf("profile %q is not logged in", profileName)
	}
	return profileName, profile, token, nil
}

func saveProfileOrganization(configPath, profileName string, organization app.Organization) error {
	if organization.ID == "" {
		return fmt.Errorf("organization id is required")
	}
	store := config.NewStore(configPath)
	cfg, err := store.Load()
	if err != nil {
		return err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]config.Profile{}
	}
	profile := cfg.Profiles[profileName]
	if profile.Domain == "" {
		profile.Domain = defaultYunxiaoDomain
	}
	profile.Organization = organization.ID
	cfg.Profiles[profileName] = profile
	if cfg.Current == "" {
		cfg.Current = profileName
	}
	return store.Save(cfg)
}
