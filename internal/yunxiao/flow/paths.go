package flow

import (
	"fmt"
	"net/url"
)

func (a *Adapter) orgPath(path string) string {
	if !a.client.IsCenter() {
		return "/oapi/v1/flow" + path
	}
	return fmt.Sprintf("/oapi/v1/flow/organizations/%s%s", url.PathEscape(a.client.OrganizationID()), path)
}

func (a *Adapter) pipelineRunsPath(pipelineID string) string {
	return a.orgPath("/pipelines/" + url.PathEscape(pipelineID) + "/runs")
}

func (a *Adapter) pipelineRunsFallbackPath(pipelineID string) string {
	return a.orgPath("/pipelines/" + url.PathEscape(pipelineID) + "/pipelineRuns")
}

func (a *Adapter) pipelineRunPath(pipelineID, runID string) string {
	return a.orgPath("/pipelines/" + url.PathEscape(pipelineID) + "/runs/" + url.PathEscape(runID))
}

func (a *Adapter) pipelineRunFallbackPath(pipelineID, runID string) string {
	return a.orgPath("/pipelines/" + url.PathEscape(pipelineID) + "/pipelineRuns/" + url.PathEscape(runID))
}

func (a *Adapter) pipelineJobStepsPath(pipelineID, runID, jobID string) string {
	return a.orgPath("/pipelines/" + url.PathEscape(pipelineID) + "/pipelineRuns/" + url.PathEscape(runID) + "/jobs/" + url.PathEscape(jobID) + "/steps")
}

func (a *Adapter) pipelineJobStepLogPath(pipelineID, runID, jobID string) string {
	return a.orgPath("/pipelines/" + url.PathEscape(pipelineID) + "/pipelineRuns/" + url.PathEscape(runID) + "/jobs/" + url.PathEscape(jobID) + "/step/log")
}

func (a *Adapter) pipelineJobRunLogFallbackPath(pipelineID, runID, jobID string) string {
	return a.orgPath("/pipelines/" + url.PathEscape(pipelineID) + "/runs/" + url.PathEscape(runID) + "/job/" + url.PathEscape(jobID) + "/log")
}
