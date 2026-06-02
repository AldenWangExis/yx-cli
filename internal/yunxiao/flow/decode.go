package flow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

type pipelineResponse struct {
	ID           flexibleString `json:"id"`
	PipelineID   flexibleString `json:"pipelineId"`
	Name         string         `json:"name"`
	PipelineName string         `json:"pipelineName"`
	Status       string         `json:"status"`
}

type pipelineRunListResponse struct {
	PipelineRuns []pipelineRunResponse `json:"pipelineRuns"`
	Runs         []pipelineRunResponse `json:"runs"`
	Items        []pipelineRunResponse `json:"items"`
	Data         []pipelineRunResponse `json:"data"`
}

type pipelineRunDetailResponse struct {
	PipelineRun pipelineRunResponse `json:"pipelineRun"`
	Run         pipelineRunResponse `json:"run"`
	Data        pipelineRunResponse `json:"data"`
}

type pipelineRunResponse struct {
	ID            flexibleString          `json:"id"`
	RunID         flexibleString          `json:"runId"`
	PipelineRunID flexibleString          `json:"pipelineRunId"`
	PipelineID    flexibleString          `json:"pipelineId"`
	Status        string                  `json:"status"`
	Branch        string                  `json:"branch"`
	Tag           string                  `json:"tag"`
	CommitID      string                  `json:"commitId"`
	Commit        string                  `json:"commit"`
	Trigger       flexibleString          `json:"trigger"`
	TriggerMode   flexibleString          `json:"triggerMode"`
	CreatedAt     flexibleString          `json:"createdAt"`
	CreateTime    flexibleString          `json:"createTime"`
	StartedAt     flexibleString          `json:"startedAt"`
	StartTime     flexibleString          `json:"startTime"`
	FinishedAt    flexibleString          `json:"finishedAt"`
	FinishTime    flexibleString          `json:"finishTime"`
	Sources       []pipelineRunSource     `json:"sources"`
	Stages        []pipelineStageResponse `json:"stages"`
	Jobs          []pipelineJobResponse   `json:"jobs"`
}

type pipelineRunSource struct {
	Data pipelineRunSourceData `json:"data"`
}

type pipelineRunSourceData struct {
	Branch   string `json:"branch"`
	Tag      string `json:"tag"`
	Commit   string `json:"commit"`
	CommitID string `json:"commitId"`
}

type pipelineStageResponse struct {
	ID        flexibleString        `json:"id"`
	Name      string                `json:"name"`
	Status    string                `json:"status"`
	Jobs      []pipelineJobResponse `json:"jobs"`
	StageInfo pipelineStageInfo     `json:"stageInfo"`
}

type pipelineStageInfo struct {
	ID     flexibleString        `json:"id"`
	Name   string                `json:"name"`
	Status string                `json:"status"`
	Jobs   []pipelineJobResponse `json:"jobs"`
}

type pipelineJobResponse struct {
	ID     flexibleString `json:"id"`
	JobID  flexibleString `json:"jobId"`
	Name   string         `json:"name"`
	Status string         `json:"status"`
}

type pipelineJobStepsResponse struct {
	ActionName        string                    `json:"actionName"`
	BuildID           flexibleString            `json:"buildId"`
	JobID             flexibleString            `json:"jobId"`
	Steps             []pipelineJobStepResponse `json:"steps"`
	Data              []pipelineJobStepResponse `json:"data"`
	BuildProcessNodes []pipelineJobStepResponse `json:"buildProcessNodes"`
}

type pipelineJobStepResponse struct {
	StepIndex flexibleString `json:"stepIndex"`
	BuildID   flexibleString `json:"buildId"`
	Name      string         `json:"name"`
	StepName  string         `json:"stepName"`
	NodeName  string         `json:"nodeName"`
	Status    string         `json:"status"`
}

type pipelineJobRunLogResponse struct {
	Content string   `json:"content"`
	Log     string   `json:"log"`
	Logs    string   `json:"logs"`
	Lines   []string `json:"lines"`
	Last    int      `json:"last"`
	More    bool     `json:"more"`
}

type pipelineCreateRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}
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
	var boolean bool
	if err := json.Unmarshal(data, &boolean); err == nil {
		if boolean {
			*s = "1"
		} else {
			*s = "0"
		}
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

func decodePipelineRunList(data []byte) ([]pipelineRunResponse, error) {
	var list []pipelineRunResponse
	listErr := json.Unmarshal(data, &list)
	if listErr == nil {
		return list, nil
	}
	var response pipelineRunListResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode pipeline runs: list %v; object %w", listErr, err)
	}
	switch {
	case response.PipelineRuns != nil:
		return response.PipelineRuns, nil
	case response.Runs != nil:
		return response.Runs, nil
	case response.Items != nil:
		return response.Items, nil
	default:
		return response.Data, nil
	}
}

func decodePipelineRunDetail(data []byte) (pipelineRunResponse, error) {
	var response pipelineRunDetailResponse
	if err := json.Unmarshal(data, &response); err == nil {
		switch {
		case response.PipelineRun.PipelineRunID != "" || response.PipelineRun.ID != "" || response.PipelineRun.RunID != "":
			return response.PipelineRun, nil
		case response.Run.PipelineRunID != "" || response.Run.ID != "" || response.Run.RunID != "":
			return response.Run, nil
		case response.Data.PipelineRunID != "" || response.Data.ID != "" || response.Data.RunID != "":
			return response.Data, nil
		}
	}
	var run pipelineRunResponse
	if err := json.Unmarshal(data, &run); err != nil {
		return pipelineRunResponse{}, fmt.Errorf("decode pipeline run: %w", err)
	}
	return run, nil
}

func decodePipelineRun(response pipelineRunResponse) app.PipelineRun {
	id := string(response.PipelineRunID)
	if id == "" {
		id = string(response.RunID)
	}
	if id == "" {
		id = string(response.ID)
	}
	branch := response.Branch
	tag := response.Tag
	commit := response.CommitID
	if commit == "" {
		commit = response.Commit
	}
	for _, source := range response.Sources {
		if branch == "" {
			branch = source.Data.Branch
		}
		if tag == "" {
			tag = source.Data.Tag
		}
		if commit == "" {
			commit = source.Data.CommitID
		}
		if commit == "" {
			commit = firstCommitID(source.Data.Commit)
		}
	}
	trigger := string(response.Trigger)
	if trigger == "" {
		trigger = string(response.TriggerMode)
	}
	createdAt := string(response.CreatedAt)
	if createdAt == "" {
		createdAt = string(response.CreateTime)
	}
	startedAt := string(response.StartedAt)
	if startedAt == "" {
		startedAt = string(response.StartTime)
	}
	finishedAt := string(response.FinishedAt)
	if finishedAt == "" {
		finishedAt = string(response.FinishTime)
	}
	stages := make([]app.PipelineStage, 0, len(response.Stages))
	jobs := make([]app.PipelineJob, 0, len(response.Jobs))
	for _, stage := range response.Stages {
		stageID := string(stage.ID)
		if stageID == "" {
			stageID = string(stage.StageInfo.ID)
		}
		stageName := stage.Name
		if stageName == "" {
			stageName = stage.StageInfo.Name
		}
		stageStatus := stage.Status
		if stageStatus == "" {
			stageStatus = stage.StageInfo.Status
		}
		stageJobs := stage.Jobs
		if len(stageJobs) == 0 {
			stageJobs = stage.StageInfo.Jobs
		}
		decodedJobs := decodePipelineJobs(stageJobs)
		stages = append(stages, app.PipelineStage{
			ID:     stageID,
			Name:   stageName,
			Status: stageStatus,
			Jobs:   decodedJobs,
		})
		jobs = append(jobs, decodedJobs...)
	}
	jobs = append(jobs, decodePipelineJobs(response.Jobs)...)
	return app.PipelineRun{
		ID:         id,
		PipelineID: string(response.PipelineID),
		Status:     response.Status,
		Branch:     branch,
		Tag:        tag,
		CommitID:   commit,
		Trigger:    trigger,
		CreatedAt:  createdAt,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Stages:     stages,
		Jobs:       jobs,
	}
}

func decodePipelineJobs(response []pipelineJobResponse) []app.PipelineJob {
	jobs := make([]app.PipelineJob, 0, len(response))
	for _, job := range response {
		id := string(job.ID)
		if id == "" {
			id = string(job.JobID)
		}
		jobs = append(jobs, app.PipelineJob{ID: id, Name: job.Name, Status: job.Status})
	}
	return jobs
}

func decodePipelineJobSteps(data []byte) ([]app.PipelineJobStep, error) {
	var groups []pipelineJobStepsResponse
	if err := json.Unmarshal(data, &groups); err == nil {
		steps := make([]app.PipelineJobStep, 0)
		for _, group := range groups {
			groupSteps := jobStepResponses(group)
			steps = append(steps, decodePipelineJobStepResponses(groupSteps, string(group.BuildID))...)
		}
		if len(steps) > 0 {
			return steps, nil
		}
	}
	var response pipelineJobStepsResponse
	if err := json.Unmarshal(data, &response); err == nil {
		stepResponses := jobStepResponses(response)
		if len(stepResponses) > 0 || response.BuildID != "" {
			return decodePipelineJobStepResponses(stepResponses, string(response.BuildID)), nil
		}
	}
	var list []pipelineJobStepResponse
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("decode pipeline job steps: %w", err)
	}
	return decodePipelineJobStepResponses(list, ""), nil
}

func jobStepResponses(response pipelineJobStepsResponse) []pipelineJobStepResponse {
	switch {
	case len(response.Steps) > 0:
		return response.Steps
	case len(response.BuildProcessNodes) > 0:
		return response.BuildProcessNodes
	default:
		return response.Data
	}
}

func decodePipelineJobStepResponses(response []pipelineJobStepResponse, fallbackBuildID string) []app.PipelineJobStep {
	steps := make([]app.PipelineJobStep, 0, len(response))
	for _, step := range response {
		buildID := string(step.BuildID)
		if buildID == "" {
			buildID = fallbackBuildID
		}
		name := step.Name
		if name == "" {
			name = step.StepName
		}
		if name == "" {
			name = step.NodeName
		}
		steps = append(steps, app.PipelineJobStep{
			StepIndex: string(step.StepIndex),
			BuildID:   buildID,
			Name:      name,
			Status:    step.Status,
		})
	}
	return steps
}

func decodePipelineJobRunLog(data []byte) (app.PipelineJobRunLog, error) {
	var response pipelineJobRunLogResponse
	if err := json.Unmarshal(data, &response); err == nil {
		content := response.Content
		if content == "" {
			content = response.Log
		}
		if content == "" {
			content = response.Logs
		}
		if content == "" && len(response.Lines) > 0 {
			content = strings.Join(response.Lines, "\n")
		}
		return app.PipelineJobRunLog{Content: content, Last: response.Last, More: response.More}, nil
	}
	var content string
	if err := json.Unmarshal(data, &content); err != nil {
		return app.PipelineJobRunLog{}, fmt.Errorf("decode pipeline job run log: %w", err)
	}
	return app.PipelineJobRunLog{Content: content}, nil
}

func firstCommitID(value string) string {
	var commits []struct {
		CommitID string `json:"commitId"`
	}
	if err := json.Unmarshal([]byte(value), &commits); err == nil && len(commits) > 0 {
		return commits[0].CommitID
	}
	return value
}
