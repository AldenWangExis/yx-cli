package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	var id flexibleString
	if err := json.Unmarshal(data, &id); err == nil && id != "" {
		return app.PipelineRun{ID: string(id), PipelineID: input.PipelineID, Status: "running"}, nil
	}
	var run app.PipelineRun
	if err := json.Unmarshal(data, &run); err != nil {
		return app.PipelineRun{}, fmt.Errorf("decode pipeline run: %w", err)
	}
	return run, nil
}

func (a *Adapter) ListPipelineRuns(ctx context.Context, input app.PipelineRunListInput) ([]app.PipelineRun, error) {
	query := url.Values{}
	if input.Branch != "" {
		query.Set("branch", input.Branch)
	}
	if input.Tag != "" {
		query.Set("tag", input.Tag)
	}
	if input.Commit != "" {
		query.Set("commit", input.Commit)
	}
	if input.Page > 0 {
		query.Set("page", strconv.Itoa(input.Page))
	}
	if input.PerPage > 0 {
		query.Set("perPage", strconv.Itoa(input.PerPage))
	}
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.pipelineRunsPath(input.PipelineID), query, nil)
	if err != nil {
		if !isNotFound(err) {
			return nil, err
		}
		data, err = a.client.DoJSON(ctx, http.MethodGet, a.pipelineRunsFallbackPath(input.PipelineID), query, nil)
		if err != nil {
			return nil, err
		}
	}
	response, err := decodePipelineRunList(data)
	if err != nil {
		return nil, err
	}
	runs := make([]app.PipelineRun, 0, len(response))
	for _, run := range response {
		runs = append(runs, decodePipelineRun(run))
	}
	return runs, nil
}

func (a *Adapter) GetPipelineRun(ctx context.Context, input app.PipelineRunGetInput) (app.PipelineRun, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.pipelineRunPath(input.PipelineID, input.RunID), nil, nil)
	if err != nil {
		if !isNotFound(err) {
			return app.PipelineRun{}, err
		}
		data, err = a.client.DoJSON(ctx, http.MethodGet, a.pipelineRunFallbackPath(input.PipelineID, input.RunID), nil, nil)
		if err != nil {
			return app.PipelineRun{}, err
		}
	}
	response, err := decodePipelineRunDetail(data)
	if err != nil {
		return app.PipelineRun{}, err
	}
	return decodePipelineRun(response), nil
}

func (a *Adapter) GetPipelineJobRunLog(ctx context.Context, input app.PipelineJobRunLogInput) (app.PipelineJobRunLog, error) {
	if input.StepIndex != "" && input.BuildID != "" {
		return a.getPipelineJobStepLog(ctx, input)
	}

	steps, err := a.GetPipelineJobSteps(ctx, input)
	if err != nil {
		if !isNotFound(err) {
			return app.PipelineJobRunLog{}, err
		}
		return a.getPipelineJobRunLogFallback(ctx, input)
	}
	if len(steps) == 0 {
		return a.getPipelineJobRunLogFallback(ctx, input)
	}
	var builder strings.Builder
	var last int
	var more bool
	for _, step := range steps {
		if step.StepIndex == "" || step.BuildID == "" {
			continue
		}
		log, err := a.getPipelineJobStepLog(ctx, app.PipelineJobRunLogInput{
			PipelineID: input.PipelineID,
			RunID:      input.RunID,
			JobID:      input.JobID,
			StepIndex:  step.StepIndex,
			BuildID:    step.BuildID,
			Offset:     input.Offset,
			Limit:      input.Limit,
		})
		if err != nil {
			return app.PipelineJobRunLog{}, err
		}
		if log.Content == "" {
			continue
		}
		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n") {
			builder.WriteByte('\n')
		}
		builder.WriteString(log.Content)
		last = log.Last
		more = more || log.More
	}
	if builder.Len() == 0 {
		return a.getPipelineJobRunLogFallback(ctx, input)
	}
	return app.PipelineJobRunLog{Content: builder.String(), Last: last, More: more}, nil
}

func (a *Adapter) GetPipelineJobSteps(ctx context.Context, input app.PipelineJobRunLogInput) ([]app.PipelineJobStep, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.pipelineJobStepsPath(input.PipelineID, input.RunID, input.JobID), nil, nil)
	if err != nil {
		return nil, err
	}
	return decodePipelineJobSteps(data)
}

func (a *Adapter) getPipelineJobStepLog(ctx context.Context, input app.PipelineJobRunLogInput) (app.PipelineJobRunLog, error) {
	query := url.Values{}
	query.Set("stepIndex", input.StepIndex)
	query.Set("buildId", input.BuildID)
	query.Set("offset", strconv.Itoa(input.Offset))
	limit := input.Limit
	if limit == 0 {
		limit = 50000
	}
	query.Set("limit", strconv.Itoa(limit))
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.pipelineJobStepLogPath(input.PipelineID, input.RunID, input.JobID), query, nil)
	if err != nil {
		return app.PipelineJobRunLog{}, err
	}
	log, err := decodePipelineJobRunLog(data)
	if err != nil {
		return app.PipelineJobRunLog{}, err
	}
	return log, nil
}

func (a *Adapter) getPipelineJobRunLogFallback(ctx context.Context, input app.PipelineJobRunLogInput) (app.PipelineJobRunLog, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.pipelineJobRunLogFallbackPath(input.PipelineID, input.RunID, input.JobID), nil, nil)
	if err != nil {
		return app.PipelineJobRunLog{}, err
	}
	log, err := decodePipelineJobRunLog(data)
	if err != nil {
		return app.PipelineJobRunLog{}, err
	}
	return log, nil
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

func isNotFound(err error) bool {
	var apiErr yunxiao.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}
