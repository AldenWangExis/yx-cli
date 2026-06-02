package cli

import (
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
)

type runtimeProfile struct {
	Name    string
	Profile config.Profile
	Token   string
	Store   *config.Store
}

func (o Options) resolveRuntimeProfile(ctx Context) (runtimeProfile, error) {
	store := config.NewStore(o.ConfigPath)
	cfg, err := store.Load()
	if err != nil {
		return runtimeProfile{}, err
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

	profile, ok := cfg.Profiles[profileName]
	if !ok {
		return runtimeProfile{}, fmt.Errorf("profile %q does not exist", profileName)
	}
	if ctx.Organization != "" {
		profile.Organization = ctx.Organization
	}
	if ctx.Domain != "" {
		profile.Domain = ctx.Domain
	}

	token, ok, err := auth.NewFileTokenStore(defaultTokenPath(o.ConfigPath)).Load(profileName)
	if err != nil {
		return runtimeProfile{}, err
	}
	if !ok {
		return runtimeProfile{}, fmt.Errorf("profile %q is not logged in", profileName)
	}

	return runtimeProfile{Name: profileName, Profile: profile, Token: token, Store: store}, nil
}
