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
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/projex"
	"github.com/spf13/cobra"
)

type WorkitemUseCase interface {
	ListProjects(ctx context.Context) ([]app.Project, error)
	CreateProject(ctx context.Context, input app.CreateProjectInput) (app.ProjectMutationResult, error)
	ListWorkitems(ctx context.Context, input app.WorkitemListInput) ([]app.WorkitemListItem, error)
	GetWorkitem(ctx context.Context, id string) (app.WorkitemDetail, error)
	CreateWorkitem(ctx context.Context, input app.CreateWorkitemInput) (app.WorkitemMutationResult, error)
	UpdateWorkitem(ctx context.Context, input app.UpdateWorkitemInput) (app.WorkitemMutationResult, error)
}

func newProjectCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Manage Yunxiao projects"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase()
			if err != nil {
				return err
			}
			projects, err := useCase.ListProjects(cmd.Context())
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(projects)
			}
			rows := make([][]string, 0, len(projects))
			for _, project := range projects {
				rows = append(rows, []string{project.ID, project.Name})
			}
			return renderer.WriteTable([]string{"ID", "NAME"}, rows)
		},
	})
	cmd.AddCommand(newProjectCreateCommand(opts))
	return cmd
}

func newProjectCreateCommand(opts Options) *cobra.Command {
	input := app.CreateProjectInput{Scope: "public"}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase()
			if err != nil {
				return err
			}
			result, err := useCase.CreateProject(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderProjectMutation(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.Name, "name", "", "project name")
	cmd.Flags().StringVar(&input.CustomCode, "custom-code", "", "project custom code")
	cmd.Flags().StringVar(&input.Scope, "scope", "public", "project scope")
	cmd.Flags().StringVar(&input.TemplateID, "template-id", "", "project template id")
	cmd.Flags().StringVar(&input.Description, "description", "", "project description")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newWorkitemCommand(opts Options, use string) *cobra.Command {
	cmd := &cobra.Command{Use: use, Short: "Manage Yunxiao work items"}
	cmd.AddCommand(newWorkitemListCommand(opts))
	cmd.AddCommand(newWorkitemViewCommand(opts))
	cmd.AddCommand(newWorkitemCreateCommand(opts))
	cmd.AddCommand(newWorkitemUpdateCommand(opts))
	return cmd
}

func newWorkitemListCommand(opts Options) *cobra.Command {
	var input app.WorkitemListInput
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work items",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase()
			if err != nil {
				return err
			}
			items, err := useCase.ListWorkitems(cmd.Context(), input)
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(items)
			}
			rows := make([][]string, 0, len(items))
			for _, item := range items {
				rows = append(rows, []string{item.ID, item.Title, item.Status, item.Type, item.ProjectID})
			}
			return renderer.WriteTable([]string{"ID", "TITLE", "STATUS", "TYPE", "PROJECT"}, rows)
		},
	}
	cmd.Flags().StringVar(&input.ProjectID, "project", "", "project id")
	cmd.Flags().StringVar(&input.Repo, "repo", "", "repository mapped to a project")
	return cmd
}

func newWorkitemViewCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "view <workitem-id>",
		Short: "View a work item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase()
			if err != nil {
				return err
			}
			item, err := useCase.GetWorkitem(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			renderer := output.NewRenderer(cmd.OutOrStdout())
			if ContextFromCommand(cmd).JSON {
				return renderer.WriteJSON(item)
			}
			return renderer.WriteTable(
				[]string{"ID", "TITLE", "STATUS", "TYPE", "PROJECT"},
				[][]string{{item.ID, item.Title, item.Status, item.Type, item.ProjectID}},
			)
		},
	}
}

func newWorkitemCreateCommand(opts Options) *cobra.Command {
	var input app.CreateWorkitemInput
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a work item",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase()
			if err != nil {
				return err
			}
			result, err := useCase.CreateWorkitem(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderWorkitemMutation(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.ProjectID, "project", "", "project id")
	cmd.Flags().StringVar(&input.Type, "type", "", "work item type")
	cmd.Flags().StringVar(&input.Title, "title", "", "work item title")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newWorkitemUpdateCommand(opts Options) *cobra.Command {
	var input app.UpdateWorkitemInput
	cmd := &cobra.Command{
		Use:   "update <workitem-id>",
		Short: "Update a work item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input.ID = args[0]
			useCase, err := opts.workitemUseCase()
			if err != nil {
				return err
			}
			result, err := useCase.UpdateWorkitem(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderWorkitemMutation(cmd, result)
		},
	}
	cmd.Flags().StringVar(&input.Status, "status", "", "new status")
	cmd.Flags().StringVar(&input.Assignee, "assignee", "", "assignee")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func renderProjectMutation(cmd *cobra.Command, result app.ProjectMutationResult) error {
	renderer := output.NewRenderer(cmd.OutOrStdout())
	if ContextFromCommand(cmd).JSON {
		return renderer.WriteJSON(result)
	}
	if result.DryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %s\n", result.Summary)
		return nil
	}
	project := result.Project
	return renderer.WriteTable([]string{"ID", "NAME", "CUSTOM_CODE"}, [][]string{{project.ID, project.Name, project.CustomCode}})
}

func renderWorkitemMutation(cmd *cobra.Command, result app.WorkitemMutationResult) error {
	renderer := output.NewRenderer(cmd.OutOrStdout())
	if ContextFromCommand(cmd).JSON {
		return renderer.WriteJSON(result)
	}
	if result.DryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %s\n", result.Summary)
		return nil
	}
	item := result.Workitem
	return renderer.WriteTable([]string{"ID", "TITLE", "STATUS"}, [][]string{{item.ID, item.Title, item.Status}})
}

func (o Options) workitemUseCase() (WorkitemUseCase, error) {
	if o.WorkitemUseCase != nil {
		return o.WorkitemUseCase, nil
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
	adapter := projex.NewAdapter(yunxiao.ClientConfig{
		BaseURL:        profile.Domain,
		Token:          token,
		OrganizationID: profile.Organization,
		Region:         profile.Region,
	})
	return app.NewWorkitemUseCase(adapter, adapter, profile.RepoProjectMap, safety.Environment{
		ConfirmWrites: profile.Safety.ConfirmWrites,
	}), nil
}
