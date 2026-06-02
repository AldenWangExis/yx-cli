package cli

import (
	"os"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/spf13/cobra"
)

func resolveRepositoryID(cmd *cobra.Command, opts Options, repo string) (string, error) {
	if repo != "" {
		return repo, nil
	}
	current, err := resolveCurrentRepository(cmd, opts)
	if err != nil {
		return "", err
	}
	return string(current.ID), nil
}

func resolveCurrentRepository(cmd *cobra.Command, opts Options) (app.CurrentRepository, error) {
	resolver, runtimeContext, err := opts.repoCurrentResolver(ContextFromCommand(cmd))
	if err != nil {
		return app.CurrentRepository{}, err
	}
	workDir, err := opts.currentWorkDir()
	if err != nil {
		return app.CurrentRepository{}, err
	}
	return resolver.CurrentRepository(cmd.Context(), app.CurrentRepositoryInput{
		ProfileName:  runtimeContext.ProfileName,
		Domain:       runtimeContext.Domain,
		Organization: runtimeContext.Organization,
		Region:       runtimeContext.Region,
		WorkDir:      workDir,
	})
}

func (o Options) currentWorkDir() (string, error) {
	if o.WorkDir != "" {
		return o.WorkDir, nil
	}
	return os.Getwd()
}
