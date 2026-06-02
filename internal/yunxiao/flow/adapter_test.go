package flow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

func TestAdapterListViewRunLogs(t *testing.T) {
	var runCalls int
	var createBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") != "token-1" {
			t.Fatalf("missing token header")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelines":
			_, _ = w.Write([]byte(`[{"pipelineId":123,"pipelineName":"Build"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelines":
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			createBody = string(body)
			_, _ = w.Write([]byte(`456`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelines/pipe1":
			_, _ = w.Write([]byte(`{"id":123,"name":"Build"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs":
			runCalls++
			_, _ = w.Write([]byte(`789`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs":
			if r.URL.Query().Get("branch") != "main" || r.URL.Query().Get("tag") != "v1.0.0-alpha" || r.URL.Query().Get("commit") != "abc123" {
				t.Fatalf("unexpected run list query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"pipelineRuns":[{"pipelineRunId":789,"pipelineId":123,"status":"SUCCESS","sources":[{"data":{"branch":"main","commit":"abc123"}}]}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs/run1":
			_, _ = w.Write([]byte(`{"pipelineRun":{"pipelineRunId":789,"pipelineId":123,"status":"SUCCESS","stages":[{"stageInfo":{"id":340,"name":"Test","status":"SUCCESS","jobs":[{"id":456,"name":"Run tests","status":"SUCCESS"}]}}]}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1/jobs/job1/steps":
			_, _ = w.Write([]byte(`{"actionName":"Go test","buildId":99,"jobId":456,"steps":[{"stepIndex":1,"stepName":"Run go test","status":"FAIL"},{"stepIndex":2,"stepName":"Collect output","status":"SUCCESS"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1/jobs/job1/step/log":
			if r.URL.Query().Get("stepIndex") == "" || r.URL.Query().Get("buildId") != "99" || r.URL.Query().Get("offset") != "0" || r.URL.Query().Get("limit") != "50000" {
				t.Fatalf("unexpected step log query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"logs":"line 1\nline 2","last":2,"more":false}`))
		case r.Method == http.MethodGet && r.URL.Path == "/oapi/v1/flow/organizations/org-1/pipelineRuns/run1/logs":
			_, _ = w.Write([]byte(`{"lines":["line 1","line 2"]}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})
	listed, err := adapter.ListPipelines(context.Background())
	if err != nil || len(listed) != 1 {
		t.Fatalf("unexpected list=%+v err=%v", listed, err)
	}
	if listed[0].ID != "123" || listed[0].Name != "Build" {
		t.Fatalf("unexpected pipeline projection: %+v", listed[0])
	}
	detail, err := adapter.GetPipeline(context.Background(), "pipe1")
	if err != nil || detail.ID != "123" {
		t.Fatalf("unexpected detail=%+v err=%v", detail, err)
	}
	created, err := adapter.CreatePipeline(context.Background(), app.PipelineCreateInput{Name: "yx-cli-ci", Content: "stages: []\n"})
	if err != nil || created.ID != "456" || created.Name != "yx-cli-ci" {
		t.Fatalf("unexpected create=%+v err=%v", created, err)
	}
	if !strings.Contains(createBody, `"name":"yx-cli-ci"`) || !strings.Contains(createBody, `"content":"stages: []\n"`) {
		t.Fatalf("unexpected create body: %s", createBody)
	}
	run, err := adapter.RunPipeline(context.Background(), app.PipelineRunInput{PipelineID: "pipe1", Branch: "main"})
	if err != nil || run.ID != "789" || run.PipelineID != "pipe1" {
		t.Fatalf("unexpected run=%+v err=%v", run, err)
	}
	if runCalls != 1 {
		t.Fatalf("expected one run call, got %d", runCalls)
	}
	runs, err := adapter.ListPipelineRuns(context.Background(), app.PipelineRunListInput{PipelineID: "pipe1", Branch: "main", Tag: "v1.0.0-alpha", Commit: "abc123"})
	if err != nil || len(runs) != 1 || runs[0].ID != "789" || runs[0].Branch != "main" {
		t.Fatalf("unexpected runs=%+v err=%v", runs, err)
	}
	runDetail, err := adapter.GetPipelineRun(context.Background(), app.PipelineRunGetInput{PipelineID: "pipe1", RunID: "run1"})
	if err != nil || len(runDetail.Jobs) != 1 || runDetail.Jobs[0].ID != "456" {
		t.Fatalf("unexpected run detail=%+v err=%v", runDetail, err)
	}
	steps, err := adapter.GetPipelineJobSteps(context.Background(), app.PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1"})
	if err != nil || len(steps) != 2 || steps[0].BuildID != "99" {
		t.Fatalf("unexpected job steps=%+v err=%v", steps, err)
	}
	jobLogs, err := adapter.GetPipelineJobRunLog(context.Background(), app.PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1"})
	if err != nil || !strings.Contains(jobLogs.Content, "line 1") {
		t.Fatalf("unexpected job logs=%+v err=%v", jobLogs, err)
	}
	logs, err := adapter.GetPipelineLogs(context.Background(), app.PipelineLogsInput{RunID: "run1", Follow: true})
	if err != nil || len(logs) != 2 {
		t.Fatalf("unexpected logs=%+v err=%v", logs, err)
	}
}

func TestAdapterListPipelineRunsFallsBackToRunsPath(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errorMessage":"not found"}`))
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns":
			_, _ = w.Write([]byte(`[{"id":789,"pipelineId":123,"status":"SUCCESS"}]`))
		default:
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})

	runs, err := adapter.ListPipelineRuns(context.Background(), app.PipelineRunListInput{PipelineID: "pipe1"})
	if err != nil {
		t.Fatalf("expected fallback list to succeed, got: %v", err)
	}
	if len(runs) != 1 || runs[0].ID != "789" {
		t.Fatalf("unexpected fallback runs: %+v", runs)
	}
	if len(paths) != 2 {
		t.Fatalf("expected primary and fallback requests, got %+v", paths)
	}
}

func TestAdapterListPipelineRunsDoesNotFallbackOnUnauthorized(t *testing.T) {
	var fallbackRequested bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"errorMessage":"unauthorized"}`))
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns":
			fallbackRequested = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})

	_, err := adapter.ListPipelineRuns(context.Background(), app.PipelineRunListInput{PipelineID: "pipe1"})
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if fallbackRequested {
		t.Fatal("expected unauthorized list response to skip fallback")
	}
}

func TestAdapterGetPipelineRunFallbacksOnlyOnNotFound(t *testing.T) {
	t.Run("falls back on not found", func(t *testing.T) {
		var fallbackRequested bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs/run1":
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errorMessage":"not found"}`))
			case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1":
				fallbackRequested = true
				_, _ = w.Write([]byte(`{"pipelineRun":{"pipelineRunId":789,"pipelineId":123,"status":"SUCCESS"}}`))
			default:
				t.Fatalf("unexpected request: %s", r.URL.Path)
			}
		}))
		defer server.Close()
		adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})

		run, err := adapter.GetPipelineRun(context.Background(), app.PipelineRunGetInput{PipelineID: "pipe1", RunID: "run1"})
		if err != nil {
			t.Fatalf("expected fallback run to succeed, got: %v", err)
		}
		if !fallbackRequested {
			t.Fatal("expected fallback run endpoint to be requested")
		}
		if run.ID != "789" || run.PipelineID != "123" || run.Status != "SUCCESS" {
			t.Fatalf("unexpected fallback run: %+v", run)
		}
	})

	t.Run("does not fall back on server error", func(t *testing.T) {
		var fallbackRequested bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs/run1":
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"errorMessage":"temporary failure"}`))
			case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1":
				fallbackRequested = true
				_, _ = w.Write([]byte(`{"pipelineRun":{"pipelineRunId":789}}`))
			default:
				t.Fatalf("unexpected request: %s", r.URL.Path)
			}
		}))
		defer server.Close()
		adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})

		_, err := adapter.GetPipelineRun(context.Background(), app.PipelineRunGetInput{PipelineID: "pipe1", RunID: "run1"})
		if err == nil {
			t.Fatal("expected server error")
		}
		if fallbackRequested {
			t.Fatal("expected server error response to skip fallback")
		}
	})
}

func TestAdapterGetPipelineJobRunLogFallsBackWhenStepsNotFound(t *testing.T) {
	var fallbackRequested bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1/jobs/job1/steps":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errorMessage":"not found"}`))
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs/run1/job/job1/log":
			fallbackRequested = true
			_, _ = w.Write([]byte(`{"content":"legacy log","last":10,"more":false}`))
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1/jobs/job1/step/log":
			t.Fatal("step log endpoint should not be requested when steps are missing")
		default:
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})

	log, err := adapter.GetPipelineJobRunLog(context.Background(), app.PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1"})
	if err != nil {
		t.Fatalf("expected fallback log to succeed, got: %v", err)
	}
	if !fallbackRequested {
		t.Fatal("expected legacy log endpoint to be requested")
	}
	if log.Content != "legacy log" || log.Last != 10 || log.More {
		t.Fatalf("unexpected fallback log: %+v", log)
	}
}

func TestAdapterGetPipelineJobRunLogDoesNotFallbackWhenStepLogFails(t *testing.T) {
	var fallbackRequested bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1/jobs/job1/steps":
			_, _ = w.Write([]byte(`{"buildId":99,"steps":[{"stepIndex":1,"stepName":"Run tests"}]}`))
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1/jobs/job1/step/log":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errorMessage":"log backend failed"}`))
		case "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/runs/run1/job/job1/log":
			fallbackRequested = true
			_, _ = w.Write([]byte(`{"content":"legacy log"}`))
		default:
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})

	_, err := adapter.GetPipelineJobRunLog(context.Background(), app.PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1"})
	if err == nil {
		t.Fatal("expected step log error")
	}
	if fallbackRequested {
		t.Fatal("expected step log error to skip legacy fallback")
	}
}

func TestAdapterDecodesBuildProcessNodesFromJobSteps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/oapi/v1/flow/organizations/org-1/pipelines/pipe1/pipelineRuns/run1/jobs/job1/steps" {
			t.Fatalf("unexpected request: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"buildId":99,"jobId":99,"buildProcessNodes":[{"stepIndex":0,"stepName":"Clone","status":"success"},{"stepIndex":3,"stepName":"Run go test","status":"fail"}]}]`))
	}))
	defer server.Close()
	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})

	steps, err := adapter.GetPipelineJobSteps(context.Background(), app.PipelineJobRunLogInput{PipelineID: "pipe1", RunID: "run1", JobID: "job1"})
	if err != nil {
		t.Fatalf("expected steps to decode, got: %v", err)
	}
	if len(steps) != 2 || steps[1].StepIndex != "3" || steps[1].BuildID != "99" || steps[1].Name != "Run go test" {
		t.Fatalf("unexpected steps: %+v", steps)
	}
}

func TestAdapterDoesNotRetryPipelineRun(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "temporary failure", http.StatusInternalServerError)
	}))
	defer server.Close()
	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "token-1", OrganizationID: "org-1", Region: "center"})
	_, err := adapter.RunPipeline(context.Background(), app.PipelineRunInput{PipelineID: "pipe1", Branch: "main"})
	if err == nil {
		t.Fatal("expected run to fail")
	}
	if calls != 1 {
		t.Fatalf("expected no retry, got %d calls", calls)
	}
}

func TestAdapterErrorsDoNotLeakToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "req-1")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"bad secret-token"}`))
	}))
	defer server.Close()
	adapter := NewAdapter(yunxiao.ClientConfig{BaseURL: server.URL, Token: "secret-token", OrganizationID: "org-1", Region: "center"})
	_, err := adapter.ListPipelines(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("leaked token: %v", err)
	}
}
