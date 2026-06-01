package cli

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/output"
	"github.com/spf13/cobra"
)

type PipelineUseCase interface {
	ListPipelines(ctx context.Context) ([]app.PipelineListItem, error)
	GetPipeline(ctx context.Context, id string) (app.PipelineDetail, error)
	RunPipeline(ctx context.Context, input app.PipelineRunInput) (app.PipelineRunResult, error)
	GetPipelineLogs(ctx context.Context, input app.PipelineLogsInput) ([]string, error)
}

func newPipelineCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{Use: "pipeline", Short: "Manage Flow pipelines"}
	cmd.AddCommand(newPipelineListCommand(opts), newPipelineViewCommand(opts), newPipelineRunCommand(opts), newPipelineLogsCommand(opts))
	return cmd
}

func newPipelineListCommand(opts Options) *cobra.Command {
	return &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
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
	return &cobra.Command{Use: "view <pipeline-id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
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

func newPipelineRunCommand(opts Options) *cobra.Command {
	var input app.PipelineRunInput
	cmd := &cobra.Command{Use: "run <pipeline-id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd := &cobra.Command{Use: "logs <run-id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
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
	return nil, fmt.Errorf("pipeline service is not configured")
}
