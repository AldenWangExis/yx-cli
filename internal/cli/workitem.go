package cli

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/spf13/cobra"
)

type WorkitemUseCase interface {
	ListProjects(ctx context.Context) ([]app.Project, error)
	CreateProject(ctx context.Context, input app.CreateProjectInput) (app.ProjectMutationResult, error)
	ListWorkitems(ctx context.Context, input app.WorkitemListInput) ([]app.WorkitemListItem, error)
	GetWorkitem(ctx context.Context, id string) (app.WorkitemDetail, error)
	CreateWorkitem(ctx context.Context, input app.CreateWorkitemInput) (app.WorkitemMutationResult, error)
	UpdateWorkitem(ctx context.Context, input app.UpdateWorkitemInput) (app.WorkitemMutationResult, error)
	DeleteWorkitem(ctx context.Context, input app.DeleteWorkitemInput) (app.WorkitemMutationResult, error)
}

func newProjectCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Manage Yunxiao projects",
		Long:    "Manage Yunxiao Projex projects.",
		Example: "  yx project list\n  yx project create --name demo --description \"Demo project\" --yes",
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List projects",
		Example: "  yx project list\n  yx --json project list",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase(ContextFromCommand(cmd))
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
		Use:     "create",
		Short:   "Create a project",
		Example: "  yx project create --name demo --dry-run\n  yx project create --name demo --description \"Demo project\" --yes",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase(ContextFromCommand(cmd))
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
	cmd := &cobra.Command{
		Use:     use,
		Short:   "Manage Yunxiao work items",
		Long:    "Manage Yunxiao work items. The issue command is an alias over workitem.",
		Example: fmt.Sprintf("  yx %s list\n  yx %s list --project <project-id>\n  yx %s view <workitem-id>\n  yx %s create --project <project-id> --type Task --title \"Do work\" --dry-run\n  yx %s update <workitem-id> --status done --dry-run\n  yx %s delete <workitem-id> --dry-run", use, use, use, use, use, use),
	}
	cmd.AddCommand(newWorkitemListCommand(opts))
	cmd.AddCommand(newWorkitemViewCommand(opts))
	cmd.AddCommand(newWorkitemCreateCommand(opts))
	cmd.AddCommand(newWorkitemUpdateCommand(opts))
	cmd.AddCommand(newWorkitemDeleteCommand(opts))
	return cmd
}

func newWorkitemListCommand(opts Options) *cobra.Command {
	var input app.WorkitemListInput
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List work items",
		Example: "  yx workitem list\n  yx workitem list --project <project-id>\n  yx issue list --repo <repo>",
		RunE: func(cmd *cobra.Command, args []string) error {
			if input.ProjectID == "" && input.Repo == "" {
				repoID, err := resolveRepositoryID(cmd, opts, "")
				if err != nil {
					return err
				}
				input.Repo = repoID
			}
			useCase, err := opts.workitemUseCase(ContextFromCommand(cmd))
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
		Use:     "view <workitem-id>",
		Short:   "View a work item",
		Example: "  yx workitem view <workitem-id>\n  yx --json workitem view <workitem-id>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase(ContextFromCommand(cmd))
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
		Use:     "create",
		Short:   "Create a work item",
		Example: "  yx workitem create --project <project-id> --type Task --title \"Do work\" --dry-run\n  yx workitem create --project <project-id> --type Task --title \"Do work\" --yes",
		RunE: func(cmd *cobra.Command, args []string) error {
			useCase, err := opts.workitemUseCase(ContextFromCommand(cmd))
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
		Use:     "update <workitem-id>",
		Short:   "Update a work item",
		Example: "  yx workitem update <workitem-id> --status done --dry-run\n  yx workitem update <workitem-id> --assignee <user-id> --yes",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input.ID = args[0]
			useCase, err := opts.workitemUseCase(ContextFromCommand(cmd))
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

func newWorkitemDeleteCommand(opts Options) *cobra.Command {
	var input app.DeleteWorkitemInput
	cmd := &cobra.Command{
		Use:     "delete <workitem-id>",
		Short:   "Delete a work item",
		Example: "  yx workitem delete <workitem-id> --dry-run\n  yx issue delete <workitem-id> --yes",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input.ID = args[0]
			useCase, err := opts.workitemUseCase(ContextFromCommand(cmd))
			if err != nil {
				return err
			}
			result, err := useCase.DeleteWorkitem(cmd.Context(), input)
			if err != nil {
				return err
			}
			return renderWorkitemMutation(cmd, result)
		},
	}
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

func (o Options) workitemUseCase(ctx Context) (WorkitemUseCase, error) {
	if o.WorkitemUseCase != nil {
		return o.WorkitemUseCase, nil
	}
	services, err := o.resolveRuntimeServices(ctx)
	if err != nil {
		return nil, err
	}
	return services.workitemUseCase(), nil
}
