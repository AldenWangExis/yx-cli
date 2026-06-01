package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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
		detail := decodePipeline(p)
		items = append(items, app.PipelineListItem{ID: detail.ID, Name: detail.Name, Status: detail.Status})
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
	return decodePipeline(p), nil
}

func (a *Adapter) CreatePipeline(ctx context.Context, input app.PipelineCreateInput) (app.PipelineDetail, error) {
	body, err := json.Marshal(pipelineCreateRequest{Name: input.Name, Content: input.Content})
	if err != nil {
		return app.PipelineDetail{}, err
	}
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.orgPath("/pipelines"), nil, body)
	if err != nil {
		return app.PipelineDetail{}, err
	}
	var id flexibleString
	if err := json.Unmarshal(data, &id); err == nil && id != "" {
		return app.PipelineDetail{ID: string(id), Name: input.Name}, nil
	}
	var p pipelineResponse
	if err := json.Unmarshal(data, &p); err != nil {
		return app.PipelineDetail{}, fmt.Errorf("decode pipeline: %w", err)
	}
	detail := decodePipeline(p)
	if detail.Name == "" {
		detail.Name = input.Name
	}
	return detail, nil
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
	ID           flexibleString `json:"id"`
	PipelineID   flexibleString `json:"pipelineId"`
	Name         string         `json:"name"`
	PipelineName string         `json:"pipelineName"`
	Status       string         `json:"status"`
}

type pipelineCreateRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err == nil {
		*s = flexibleString(value)
		return nil
	}
	var number int64
	if err := json.Unmarshal(data, &number); err == nil {
		*s = flexibleString(strconv.FormatInt(number, 10))
		return nil
	}
	return fmt.Errorf("decode string: %s", string(data))
}

func decodePipeline(response pipelineResponse) app.PipelineDetail {
	id := string(response.ID)
	if id == "" {
		id = string(response.PipelineID)
	}
	name := response.Name
	if name == "" {
		name = response.PipelineName
	}
	return app.PipelineDetail{ID: id, Name: name, Status: response.Status}
}
