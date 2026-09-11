//go:build unit

package sandbox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSandboxControlPlaneDelegation(t *testing.T) {
	tests := []struct {
		name        string
		sandboxCall func(context.Context, *Sandbox) (any, error)
		clientCall  func(context.Context, *Client) (any, error)
	}{
		{name: "GetInfo", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return sb.GetInfo(ctx) }, clientCall: func(ctx context.Context, c *Client) (any, error) { return c.GetInfo(ctx, "sandbox-123") }},
		{name: "GetMetrics", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return sb.GetMetrics(ctx, nil) }, clientCall: func(ctx context.Context, c *Client) (any, error) {
			return c.GetMetrics(ctx, "sandbox-123", nil)
		}},
		{name: "GetLogs", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return sb.GetLogs(ctx, nil) }, clientCall: func(ctx context.Context, c *Client) (any, error) { return c.GetLogs(ctx, "sandbox-123", nil) }},
		{name: "GetResources", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return sb.GetResources(ctx) }, clientCall: func(ctx context.Context, c *Client) (any, error) { return c.GetResources(ctx, "sandbox-123") }},
		{name: "GetInjections", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return sb.GetInjections(ctx) }, clientCall: func(ctx context.Context, c *Client) (any, error) { return c.GetInjections(ctx, "sandbox-123") }},
		{name: "Kill", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return nil, sb.Kill(ctx) }, clientCall: func(ctx context.Context, c *Client) (any, error) { return nil, c.Kill(ctx, "sandbox-123") }},
		{name: "Pause", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return nil, sb.Pause(ctx) }, clientCall: func(ctx context.Context, c *Client) (any, error) { return nil, c.Pause(ctx, "sandbox-123") }},
		{name: "Refresh", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return nil, sb.Refresh(ctx, RefreshParams{}) }, clientCall: func(ctx context.Context, c *Client) (any, error) {
			return nil, c.Refresh(ctx, "sandbox-123", RefreshParams{})
		}},
		{name: "SetTimeout", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return nil, sb.SetTimeout(ctx, time.Minute) }, clientCall: func(ctx context.Context, c *Client) (any, error) {
			return nil, c.SetTimeout(ctx, "sandbox-123", time.Minute)
		}},
		{name: "UpdateInjections", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) { return nil, sb.UpdateInjections(ctx, nil) }, clientCall: func(ctx context.Context, c *Client) (any, error) {
			return nil, c.UpdateInjections(ctx, "sandbox-123", nil)
		}},
		{name: "UpdateGitHubToken", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) {
			return nil, sb.UpdateGitHubToken(ctx, "new-token")
		}, clientCall: func(ctx context.Context, c *Client) (any, error) {
			return nil, c.UpdateGitHubToken(ctx, "sandbox-123", "new-token")
		}},
		{name: "UpdateGitRepositoryResourceToken", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) {
			return nil, sb.UpdateGitRepositoryResourceToken(ctx, "res_123", "new-token")
		}, clientCall: func(ctx context.Context, c *Client) (any, error) {
			return nil, c.UpdateGitRepositoryResourceToken(ctx, "sandbox-123", "res_123", "new-token")
		}},
		{name: "WaitForReady", sandboxCall: func(ctx context.Context, sb *Sandbox) (any, error) {
			return sb.WaitForReady(ctx, WithPollInterval(time.Millisecond))
		}, clientCall: func(ctx context.Context, c *Client) (any, error) {
			return c.WaitForReady(ctx, "sandbox-123", WithPollInterval(time.Millisecond))
		}},
	}
	for _, tt := range tests {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/error=%t", tt.name, fail), func(t *testing.T) {
				var requests []string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body strings.Builder
					_, _ = io.Copy(&body, r.Body)
					requests = append(requests, r.Method+" "+r.URL.RequestURI()+" "+body.String())
					w.Header().Set("Content-Type", "application/json")
					if fail {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"message":"test error"}`))
						return
					}
					switch tt.name {
					case "GetInfo", "WaitForReady":
						_, _ = w.Write([]byte(`{"sandboxID":"sandbox-123","state":"running"}`))
					case "GetMetrics":
						_, _ = w.Write([]byte(`[]`))
					case "GetLogs":
						_, _ = w.Write([]byte(`{"logs":[]}`))
					case "GetResources":
						_, _ = w.Write([]byte(`{"resources":[]}`))
					case "GetInjections":
						_, _ = w.Write([]byte(`{"injections":[]}`))
					default:
						w.WriteHeader(http.StatusNoContent)
					}
				}))
				defer server.Close()
				c, err := NewClient(&Config{APIKey: "test-key", Endpoint: server.URL})
				if err != nil {
					t.Fatal(err)
				}
				sb := &Sandbox{sandboxID: "sandbox-123", client: c}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				clientResult, clientErr := tt.clientCall(ctx, c)
				sandboxResult, sandboxErr := tt.sandboxCall(ctx, sb)
				if fail && (clientErr == nil || sandboxErr == nil || clientErr.Error() != sandboxErr.Error()) {
					t.Fatalf("errors differ: client=%v sandbox=%v", clientErr, sandboxErr)
				}
				if !fail && (clientErr != nil || sandboxErr != nil || !reflect.DeepEqual(clientResult, sandboxResult)) {
					t.Fatalf("results differ: client=%#v (%v) sandbox=%#v (%v)", clientResult, clientErr, sandboxResult, sandboxErr)
				}
				if len(requests) != 2 || requests[0] != requests[1] {
					t.Fatalf("requests differ: %v", requests)
				}
				if !strings.Contains(requests[0], "/sandboxes/sandbox-123") {
					t.Fatalf("unexpected sandbox target: %v", requests)
				}
			})
		}
	}
}
