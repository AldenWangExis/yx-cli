package cli

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/AldenWangExis/yx-cli/internal/safety"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/codeup"
	"github.com/spf13/cobra"
)

type MergeRequestUseCase interface {
	ListMergeRequests(ctx context.Context, repo string) ([]app.MergeRequestListItem, error)
	GetMergeRequest(ctx context.Context, repo, id string) (app.MergeRequestDetail, error)
	CreateMergeRequest(ctx context.Context, input app.CreateMergeRequestInput) (app.MergeRequestMutationResult, error)
	MergeMergeRequest(ctx context.Context, input app.MergeMergeRequestInput) (app.MergeRequestMutationResult, error)
}

func newMergeRequestCommand(opts Options, use string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Short:   "Manage Codeup merge requests",
		Long:    "Manage Codeup merge requests for a repository.",
		Example: fmt.Sprintf("  yx %s list --repo <repo>\n  yx %s view <mr-id> --repo <repo>\n  yx %s create --repo <repo> --source feat/a --target master --title \"Add feature\" --dry-run\n  yx %s merge <mr-id> --repo <repo> --yes", use, use, use, use),
	}
	cmd.AddCommand(newMRListCommand(opts))
	cmd.AddCommand(newMRViewCommand(opts))
	cmd.AddCommand(newMRCreateCommand(opts))
	cmd.AddCommand(newMRMergeCommand(opts))
	return cmd
}

func newMRListCommand(opts Options) *cobra.Command {
	var repo string
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List merge requests",
		Example: "  yx mr list --repo 6925595\n  yx --json mr list --repo 6925595",
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				return fmt.Errorf("--repo is required")
			}
			useCase, err := opts.mergeRequestUseCase()
			if err != nil {
				return err
			}
			mrs, err := useCase.ListMergeRequests(cmd.Context(), repo)
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(mrs)
			}
			rows := make([][]string, 0, len(mrs))
			for _, mr := range mrs {
				rows = append(rows, []string{mr.ID, mr.Title, mr.State, mr.SourceBranch, mr.TargetBranch})
			}
			return renderer.WriteTable([]string{"ID", "TITLE", "STATE", "SOURCE", "TARGET"}, rows)
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "repository identifier")
	return cmd
}

func newMRViewCommand(opts Options) *cobra.Command {
	var repo string
	cmd := &cobra.Command{
		Use:     "view <mr-id>",
		Short:   "View a merge request",
		Example: "  yx mr view 1 --repo 6925595\n  yx --json mr view 1 --repo 6925595",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				return fmt.Errorf("--repo is required")
			}
			useCase, err := opts.mergeRequestUseCase()
			if err != nil {
				return err
			}
			mr, err := useCase.GetMergeRequest(cmd.Context(), repo, args[0])
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(mr)
			}
			return renderer.WriteTable(
				[]string{"ID", "TITLE", "STATE", "SOURCE", "TARGET"},
				[][]string{{mr.ID, mr.Title, mr.State, mr.SourceBranch, mr.TargetBranch}},
			)
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "repository identifier")
	return cmd
}

func newMRCreateCommand(opts Options) *cobra.Command {
	var input app.CreateMergeRequestInput
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a merge request",
		Example: "  yx mr create --repo 6925595 --source feat/a --target master --title \"Add feature\" --dry-run\n  yx mr create --repo 6925595 --source feat/a --target master --title \"Add feature\" --yes",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.mergeRequestUseCase()
			if err != nil {
				return err
			}
			result, err := useCase.CreateMergeRequest(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderMutationResult(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.Repo, "repo", "", "repository identifier")
	cmd.Flags().StringVar(&input.SourceBranch, "source", "", "source branch")
	cmd.Flags().StringVar(&input.TargetBranch, "target", "", "target branch")
	cmd.Flags().StringVar(&input.Title, "title", "", "merge request title")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newMRMergeCommand(opts Options) *cobra.Command {
	var input app.MergeMergeRequestInput
	cmd := &cobra.Command{
		Use:     "merge <mr-id>",
		Short:   "Merge a merge request",
		Example: "  yx mr merge 1 --repo 6925595 --dry-run\n  yx mr merge 1 --repo 6925595 --yes",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input.ID = args[0]
			useCase, err := opts.mergeRequestUseCase()
			if err != nil {
				return err
			}
			result, err := useCase.MergeMergeRequest(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderMutationResult(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.Repo, "repo", "", "repository identifier")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func renderMutationResult(cmd *cobra.Command, result app.MergeRequestMutationResult) error {
	renderer := output.NewRenderer(cmd.OutOrStdout())
	if ContextFromCommand(cmd).JSON {
		return renderer.WriteJSON(result)
	}
	if result.DryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %s\n", result.Summary)
		return nil
	}
	mr := result.MergeRequest
	return renderer.WriteTable(
		[]string{"ID", "TITLE", "STATE"},
		[][]string{{mr.ID, mr.Title, mr.State}},
	)
}

func (o Options) mergeRequestUseCase() (MergeRequestUseCase, error) {
	if o.MergeRequestUseCase != nil {
		return o.MergeRequestUseCase, nil
	}
	cfg, err := config.NewStore(o.ConfigPath).Load()
	if err != nil {
		return nil, err
	}
	profileName := o.DefaultProfile
	if profileName == "" {
		profileName = cfg.Current
	}
	if profileName == "" {
		profileName = "default"
	}
	profile, ok := cfg.Profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile %q does not exist", profileName)
	}
	token, ok, err := auth.NewFileTokenStore(defaultTokenPath(o.ConfigPath)).Load(profileName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("profile %q is not logged in", profileName)
	}
	service := codeup.NewChangeRequestAdapter(yunxiao.ClientConfig{
		BaseURL:        profile.Domain,
		Token:          token,
		OrganizationID: profile.Organization,
		Region:         profile.Region,
	})
	return app.NewMergeRequestUseCase(service, safety.Environment{
		ConfirmWrites: profile.Safety.ConfirmWrites,
	}), nil
}
