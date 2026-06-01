package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/auth"
	"github.com/AldenWangExis/yx-cli/internal/config"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/AldenWangExis/yx-cli/internal/safety"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/flow"
	"github.com/spf13/cobra"
)

type PipelineUseCase interface {
	ListPipelines(ctx context.Context) ([]app.PipelineListItem, error)
	GetPipeline(ctx context.Context, id string) (app.PipelineDetail, error)
	CreatePipeline(ctx context.Context, input app.PipelineCreateInput) (app.PipelineMutationResult, error)
	RunPipeline(ctx context.Context, input app.PipelineRunInput) (app.PipelineRunResult, error)
	GetPipelineLogs(ctx context.Context, input app.PipelineLogsInput) ([]string, error)
}

func newPipelineCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pipeline",
		Short:   "Manage Flow pipelines",
		Long:    "Manage Yunxiao Flow pipelines, runs, and logs.",
		Example: "  yx pipeline list\n  yx pipeline view <pipeline-id>\n  yx pipeline run <pipeline-id> --branch master --dry-run\n  yx pipeline logs <run-id>",
	}
	cmd.AddCommand(newPipelineListCommand(opts), newPipelineViewCommand(opts), newPipelineCreateCommand(opts), newPipelineRunCommand(opts), newPipelineLogsCommand(opts))
	return cmd
}

func newPipelineListCommand(opts Options) *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List pipelines", Example: "  yx pipeline list\n  yx --json pipeline list", RunE: func(cmd *cobra.Command, args []string) error {
		useCase, err := opts.pipelineUseCase()
		if err != nil {
			return err
		}
		items, err := useCase.ListPipelines(cmd.Context())
		if err != nil {
			return err
		}
		r := output.NewRenderer(cmd.OutOrStdout())
		if ContextFromCommand(cmd).JSON {
			return r.WriteJSON(items)
		}
		rows := make([][]string, 0, len(items))
		for _, item := range items {
			rows = append(rows, []string{item.ID, item.Name, item.Status})
		}
		return r.WriteTable([]string{"ID", "NAME", "STATUS"}, rows)
	}}
}

func newPipelineViewCommand(opts Options) *cobra.Command {
	return &cobra.Command{Use: "view <pipeline-id>", Short: "View a pipeline", Example: "  yx pipeline view <pipeline-id>\n  yx --json pipeline view <pipeline-id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		useCase, err := opts.pipelineUseCase()
		if err != nil {
			return err
		}
		item, err := useCase.GetPipeline(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		r := output.NewRenderer(cmd.OutOrStdout())
		if ContextFromCommand(cmd).JSON {
			return r.WriteJSON(item)
		}
		return r.WriteTable([]string{"ID", "NAME", "STATUS"}, [][]string{{item.ID, item.Name, item.Status}})
	}}
}

func newPipelineCreateCommand(opts Options) *cobra.Command {
	var input app.PipelineCreateInput
	var filePath string
	cmd := &cobra.Command{Use: "create", Short: "Create a pipeline", Example: "  yx pipeline create --name yx-cli-ci --file flow.yml --dry-run\n  yx pipeline create --name yx-cli-ci --file flow.yml --yes", RunE: func(cmd *cobra.Command, args []string) error {
		if filePath == "" {
			return fmt.Errorf("--file is required")
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read pipeline file: %w", err)
		}
		input.Content = string(data)
		useCase, err := opts.pipelineUseCase()
		if err != nil {
			return err
		}
		result, err := useCase.CreatePipeline(cmd.Context(), input)
		if err != nil {
			return err
		}
		r := output.NewRenderer(cmd.OutOrStdout())
		if ContextFromCommand(cmd).JSON {
			return r.WriteJSON(result)
		}
		if result.DryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %s\n", result.Summary)
			return nil
		}
		return r.WriteTable([]string{"ID", "NAME", "STATUS"}, [][]string{{result.Pipeline.ID, result.Pipeline.Name, result.Pipeline.Status}})
	}}
	cmd.Flags().StringVar(&input.Name, "name", "", "pipeline name")
	cmd.Flags().StringVar(&filePath, "file", "", "pipeline YAML file")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newPipelineRunCommand(opts Options) *cobra.Command {
	var input app.PipelineRunInput
	cmd := &cobra.Command{Use: "run <pipeline-id>", Short: "Run a pipeline", Example: "  yx pipeline run <pipeline-id> --branch master --dry-run\n  yx pipeline run <pipeline-id> --branch master --yes", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		input.PipelineID = args[0]
		useCase, err := opts.pipelineUseCase()
		if err != nil {
			return err
		}
		result, err := useCase.RunPipeline(cmd.Context(), input)
		if err != nil {
			return err
		}
		r := output.NewRenderer(cmd.OutOrStdout())
		if ContextFromCommand(cmd).JSON {
			return r.WriteJSON(result)
		}
		if result.DryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %s\n", result.Summary)
			return nil
		}
		return r.WriteTable([]string{"ID", "PIPELINE", "STATUS"}, [][]string{{result.Run.ID, result.Run.PipelineID, result.Run.Status}})
	}}
	cmd.Flags().StringVar(&input.Branch, "branch", "", "branch to run")
	cmd.Flags().BoolVar(&input.DryRun, "dry-run", false, "show intended operation without writing")
	cmd.Flags().BoolVar(&input.Yes, "yes", false, "skip confirmation")
	return cmd
}

func newPipelineLogsCommand(opts Options) *cobra.Command {
	var input app.PipelineLogsInput
	cmd := &cobra.Command{Use: "logs <run-id>", Short: "View pipeline run logs", Example: "  yx pipeline logs <run-id>\n  yx pipeline logs <run-id> --follow", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		input.RunID = args[0]
		useCase, err := opts.pipelineUseCase()
		if err != nil {
			return err
		}
		lines, err := useCase.GetPipelineLogs(cmd.Context(), input)
		if err != nil {
			return err
		}
		for _, line := range lines {
			fmt.Fprintln(cmd.OutOrStdout(), line)
		}
		return nil
	}}
	cmd.Flags().BoolVar(&input.Follow, "follow", false, "follow logs")
	return cmd
}

func (o Options) pipelineUseCase() (PipelineUseCase, error) {
	if o.PipelineUseCase != nil {
		return o.PipelineUseCase, nil
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
	adapter := flow.NewAdapter(yunxiao.ClientConfig{BaseURL: profile.Domain, Token: token, OrganizationID: profile.Organization, Region: profile.Region})
	return app.NewPipelineUseCase(adapter, safety.Environment{ConfirmWrites: profile.Safety.ConfirmWrites}), nil
}
