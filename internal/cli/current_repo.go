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
	return current.ID, nil
}

func resolveCurrentRepository(cmd *cobra.Command, opts Options) (app.CurrentRepository, error) {
	resolver, profileName, organization, err := opts.repoCurrentResolver(ContextFromCommand(cmd))
	if err != nil {
		return app.CurrentRepository{}, err
	}
	workDir, err := currentWorkDir()
	if err != nil {
		return app.CurrentRepository{}, err
	}
	return resolver.CurrentRepository(cmd.Context(), app.CurrentRepositoryInput{
		ProfileName:  profileName,
		Organization: organization,
		WorkDir:      workDir,
	})
}

func currentWorkDir() (string, error) {
	return os.Getwd()
}
