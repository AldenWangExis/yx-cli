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
	logs, err := adapter.GetPipelineLogs(context.Background(), app.PipelineLogsInput{RunID: "run1", Follow: true})
	if err != nil || len(logs) != 2 {
		t.Fatalf("unexpected logs=%+v err=%v", logs, err)
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
