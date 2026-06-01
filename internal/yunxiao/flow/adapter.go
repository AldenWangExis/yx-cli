package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

type Adapter struct{ client *yunxiao.Client }

func NewAdapter(config yunxiao.ClientConfig) *Adapter {
	return &Adapter{client: yunxiao.NewClient(config)}
}

func (a *Adapter) ListPipelines(ctx context.Context) ([]app.PipelineListItem, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.orgPath("/pipelines"), nil, nil)
	if err != nil {
		return nil, err
	}
	var response []pipelineResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode pipelines: %w", err)
	}
	items := make([]app.PipelineListItem, 0, len(response))
	for _, p := range response {
		items = append(items, app.PipelineListItem{ID: p.ID, Name: p.Name, Status: p.Status})
	}
	return items, nil
}

func (a *Adapter) GetPipeline(ctx context.Context, id string) (app.PipelineDetail, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.orgPath("/pipelines/"+url.PathEscape(id)), nil, nil)
	if err != nil {
		return app.PipelineDetail{}, err
	}
	var p pipelineResponse
	if err := json.Unmarshal(data, &p); err != nil {
		return app.PipelineDetail{}, fmt.Errorf("decode pipeline: %w", err)
	}
	return app.PipelineDetail{ID: p.ID, Name: p.Name, Status: p.Status}, nil
}

func (a *Adapter) RunPipeline(ctx context.Context, input app.PipelineRunInput) (app.PipelineRun, error) {
	body, _ := json.Marshal(map[string]string{"branch": input.Branch})
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.orgPath("/pipelines/"+url.PathEscape(input.PipelineID)+"/runs"), nil, body)
	if err != nil {
		return app.PipelineRun{}, err
	}
	var run app.PipelineRun
	if err := json.Unmarshal(data, &run); err != nil {
		return app.PipelineRun{}, fmt.Errorf("decode pipeline run: %w", err)
	}
	return run, nil
}

func (a *Adapter) GetPipelineLogs(ctx context.Context, input app.PipelineLogsInput) ([]string, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.orgPath("/pipelineRuns/"+url.PathEscape(input.RunID)+"/logs"), nil, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Lines []string `json:"lines"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode pipeline logs: %w", err)
	}
	return response.Lines, nil
}

func (a *Adapter) orgPath(path string) string {
	return fmt.Sprintf("/oapi/v1/flow/organizations/%s%s", url.PathEscape(a.client.OrganizationID()), path)
}

type pipelineResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
