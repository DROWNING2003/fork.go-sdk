//go:build unit

package sandbox

import (
	"context"
	"testing"
	"time"

	"github.com/qiniu/go-sdk/v7/sandbox/internal/apis"
)

func TestClientGetInfo(t *testing.T) {
	mock := &mockAPI{
		getSandboxFn: func(_ context.Context, sandboxID apis.SandboxID, _ ...apis.RequestEditorFn) (*apis.GetSandboxResponse, error) {
			if sandboxID != "sandbox-1" {
				t.Fatalf("unexpected sandbox ID: %s", sandboxID)
			}
			return &apis.GetSandboxResponse{
				JSON200:      &apis.SandboxDetail{SandboxID: sandboxID, State: apis.Running},
				HTTPResponse: httpResponse(200),
			}, nil
		},
	}

	info, err := newTestClient(mock).GetInfo(context.Background(), "sandbox-1")
	if err != nil {
		t.Fatal(err)
	}
	if info.SandboxID != "sandbox-1" || info.State != StateRunning {
		t.Fatalf("unexpected sandbox info: %#v", info)
	}
}

func TestClientGetMetrics(t *testing.T) {
	mock := &mockAPI{
		getSandboxMetricsFn: func(_ context.Context, sandboxID apis.SandboxID, params *apis.GetSandboxMetricsParams, _ ...apis.RequestEditorFn) (*apis.GetSandboxMetricsResponse, error) {
			if sandboxID != "sandbox-1" || params == nil || params.Start == nil || *params.Start != 123 {
				t.Fatalf("unexpected metrics request: %s, %#v", sandboxID, params)
			}
			metrics := []apis.SandboxMetric{{CPUCount: 2, CPUUsedPct: 50}}
			return &apis.GetSandboxMetricsResponse{
				JSON200:      &metrics,
				HTTPResponse: httpResponse(200),
			}, nil
		},
	}
	start := int64(123)
	metrics, err := newTestClient(mock).GetMetrics(context.Background(), "sandbox-1", &GetMetricsParams{Start: &start})
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 1 || metrics[0].CPUUsedPct != 50 {
		t.Fatalf("unexpected metrics: %#v", metrics)
	}
}

func TestClientGetLogs(t *testing.T) {
	mock := &mockAPI{
		getSandboxLogsFn: func(_ context.Context, sandboxID apis.SandboxID, params *apis.GetSandboxLogsParams, _ ...apis.RequestEditorFn) (*apis.GetSandboxLogsResponse, error) {
			if sandboxID != "sandbox-1" || params == nil || params.Limit == nil || *params.Limit != 10 {
				t.Fatalf("unexpected logs request: %s, %#v", sandboxID, params)
			}
			return &apis.GetSandboxLogsResponse{
				JSON200:      &apis.SandboxLogs{Logs: []apis.SandboxLog{{Line: "hello"}}},
				HTTPResponse: httpResponse(200),
			}, nil
		},
	}
	limit := int32(10)
	logs, err := newTestClient(mock).GetLogs(context.Background(), "sandbox-1", &GetLogsParams{Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs.Logs) != 1 || logs.Logs[0].Line != "hello" {
		t.Fatalf("unexpected logs: %#v", logs)
	}
}

func TestClientKill(t *testing.T) {
	mock := &mockAPI{
		deleteSandboxFn: func(_ context.Context, sandboxID apis.SandboxID, _ ...apis.RequestEditorFn) (*apis.DeleteSandboxResponse, error) {
			if sandboxID != "sandbox-1" {
				t.Fatalf("unexpected sandbox ID: %s", sandboxID)
			}
			return &apis.DeleteSandboxResponse{HTTPResponse: httpResponse(204)}, nil
		},
	}

	if err := newTestClient(mock).Kill(context.Background(), "sandbox-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClientPause(t *testing.T) {
	mock := &mockAPI{
		pauseSandboxFn: func(_ context.Context, sandboxID apis.SandboxID, _ ...apis.RequestEditorFn) (*apis.PauseSandboxResponse, error) {
			if sandboxID != "sandbox-1" {
				t.Fatalf("unexpected sandbox ID: %s", sandboxID)
			}
			return &apis.PauseSandboxResponse{HTTPResponse: httpResponse(204)}, nil
		},
	}

	if err := newTestClient(mock).Pause(context.Background(), "sandbox-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClientGetInjections(t *testing.T) {
	var injection apis.SandboxInjection
	if err := injection.FromInjectionByID(apis.InjectionByID{ID: "rule-1", Type: apis.ID}); err != nil {
		t.Fatal(err)
	}
	mock := &mockAPI{
		getSandboxInjectionsFn: func(_ context.Context, sandboxID apis.SandboxID, _ ...apis.RequestEditorFn) (*apis.GetSandboxInjectionsResponse, error) {
			if sandboxID != "sandbox-1" {
				t.Fatalf("unexpected sandbox ID: %s", sandboxID)
			}
			return &apis.GetSandboxInjectionsResponse{
				JSON200: &struct {
					Injections []apis.SandboxInjection `json:"injections"`
				}{Injections: []apis.SandboxInjection{injection}},
				HTTPResponse: httpResponse(200),
			}, nil
		},
	}

	injections, err := newTestClient(mock).GetInjections(context.Background(), "sandbox-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(injections) != 1 || injections[0].ByID == nil || *injections[0].ByID != "rule-1" {
		t.Fatalf("unexpected injections: %#v", injections)
	}
}

func TestClientUpdateInjections(t *testing.T) {
	ruleID := "rule-1"
	mock := &mockAPI{
		updateSandboxInjectionsFn: func(_ context.Context, sandboxID apis.SandboxID, body apis.UpdateSandboxInjectionsJSONRequestBody, _ ...apis.RequestEditorFn) (*apis.UpdateSandboxInjectionsResponse, error) {
			if sandboxID != "sandbox-1" || len(body.Injections) != 1 {
				t.Fatalf("unexpected request: %s, %#v", sandboxID, body.Injections)
			}
			return &apis.UpdateSandboxInjectionsResponse{HTTPResponse: httpResponse(204)}, nil
		},
	}

	if err := newTestClient(mock).UpdateInjections(context.Background(), "sandbox-1", []SandboxInjectionSpec{{ByID: &ruleID}}); err != nil {
		t.Fatal(err)
	}
}

func TestClientUpdateGitHubToken(t *testing.T) {
	mock := &mockAPI{
		updateSandboxGithubTokenFn: func(_ context.Context, sandboxID apis.SandboxID, body apis.UpdateSandboxGithubTokenJSONRequestBody, _ ...apis.RequestEditorFn) (*apis.UpdateSandboxGithubTokenResponse, error) {
			if sandboxID != "sandbox-1" || body.AuthorizationToken != "github-token" {
				t.Fatalf("unexpected request: %s, %#v", sandboxID, body)
			}
			return &apis.UpdateSandboxGithubTokenResponse{HTTPResponse: httpResponse(204)}, nil
		},
	}

	if err := newTestClient(mock).UpdateGitHubToken(context.Background(), "sandbox-1", "github-token"); err != nil {
		t.Fatal(err)
	}
}

func TestClientSetTimeout(t *testing.T) {
	mock := &mockAPI{
		updateSandboxTimeoutFn: func(_ context.Context, sandboxID apis.SandboxID, body apis.UpdateSandboxTimeoutJSONRequestBody, _ ...apis.RequestEditorFn) (*apis.UpdateSandboxTimeoutResponse, error) {
			if sandboxID != "sandbox-1" || body.Timeout != 30 {
				t.Fatalf("unexpected request: %s, %#v", sandboxID, body)
			}
			return &apis.UpdateSandboxTimeoutResponse{HTTPResponse: httpResponse(204)}, nil
		},
	}

	if err := newTestClient(mock).SetTimeout(context.Background(), "sandbox-1", 30*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestClientRefresh(t *testing.T) {
	duration := 60
	mock := &mockAPI{
		refreshSandboxFn: func(_ context.Context, sandboxID apis.SandboxID, body apis.RefreshSandboxJSONRequestBody, _ ...apis.RequestEditorFn) (*apis.RefreshSandboxResponse, error) {
			if sandboxID != "sandbox-1" || body.Duration == nil || *body.Duration != duration {
				t.Fatalf("unexpected request: %s, %#v", sandboxID, body)
			}
			return &apis.RefreshSandboxResponse{HTTPResponse: httpResponse(204)}, nil
		},
	}

	if err := newTestClient(mock).Refresh(context.Background(), "sandbox-1", RefreshParams{Duration: &duration}); err != nil {
		t.Fatal(err)
	}
}

func TestClientWaitForReady(t *testing.T) {
	mock := &mockAPI{
		getSandboxFn: func(_ context.Context, sandboxID apis.SandboxID, _ ...apis.RequestEditorFn) (*apis.GetSandboxResponse, error) {
			return &apis.GetSandboxResponse{
				JSON200:      &apis.SandboxDetail{SandboxID: sandboxID, State: apis.Running},
				HTTPResponse: httpResponse(200),
			}, nil
		},
	}

	info, err := newTestClient(mock).WaitForReady(context.Background(), "sandbox-1", WithPollInterval(0))
	if err != nil {
		t.Fatal(err)
	}
	if info.SandboxID != "sandbox-1" || info.State != StateRunning {
		t.Fatalf("unexpected sandbox info: %#v", info)
	}
}
