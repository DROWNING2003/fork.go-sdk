package sandbox

import (
	"context"
	"time"
)

// GetInfo 返回当前沙箱的详细信息。
// 具体行为参见 [Client.GetInfo]。
func (s *Sandbox) GetInfo(ctx context.Context) (*SandboxInfo, error) {
	return s.client.GetInfo(ctx, s.sandboxID)
}

// GetMetrics 返回当前沙箱的资源指标。
// 具体行为参见 [Client.GetMetrics]。
func (s *Sandbox) GetMetrics(ctx context.Context, params *GetMetricsParams) ([]SandboxMetric, error) {
	return s.client.GetMetrics(ctx, s.sandboxID, params)
}

// GetLogs 返回当前沙箱的日志。
// 具体行为参见 [Client.GetLogs]。
func (s *Sandbox) GetLogs(ctx context.Context, params *GetLogsParams) (*SandboxLogs, error) {
	return s.client.GetLogs(ctx, s.sandboxID, params)
}

// GetResources 返回当前沙箱已挂载的资源配置。
// 具体行为参见 [Client.GetResources]。
func (s *Sandbox) GetResources(ctx context.Context) ([]SandboxResourceInfo, error) {
	return s.client.GetResources(ctx, s.sandboxID)
}

// GetInjections 返回当前沙箱的运行时请求注入规则。
// 具体行为参见 [Client.GetInjections]。
func (s *Sandbox) GetInjections(ctx context.Context) ([]MaskedSandboxInjection, error) {
	return s.client.GetInjections(ctx, s.sandboxID)
}

// Kill 终止当前沙箱。
// 具体行为参见 [Client.Kill]。
func (s *Sandbox) Kill(ctx context.Context) error {
	return s.client.Kill(ctx, s.sandboxID)
}

// Pause 暂停当前沙箱。
// 具体行为参见 [Client.Pause]。
func (s *Sandbox) Pause(ctx context.Context) error {
	return s.client.Pause(ctx, s.sandboxID)
}

// Refresh 延长当前沙箱的存活时间。
// 具体行为参见 [Client.Refresh]。
func (s *Sandbox) Refresh(ctx context.Context, params RefreshParams) error {
	return s.client.Refresh(ctx, s.sandboxID, params)
}

// SetTimeout 更新当前沙箱的超时时间，timeout 必须至少为 1 秒。
// 具体行为参见 [Client.SetTimeout]。
func (s *Sandbox) SetTimeout(ctx context.Context, timeout time.Duration) error {
	return s.client.SetTimeout(ctx, s.sandboxID, timeout)
}

// UpdateInjections 替换当前沙箱的全部运行时请求注入规则。
// 具体行为参见 [Client.UpdateInjections]。
func (s *Sandbox) UpdateInjections(ctx context.Context, injections []SandboxInjectionSpec) error {
	return s.client.UpdateInjections(ctx, s.sandboxID, injections)
}

// UpdateGitHubToken 更新当前沙箱使用的 GitHub 授权令牌。
// 具体行为参见 [Client.UpdateGitHubToken]。
func (s *Sandbox) UpdateGitHubToken(ctx context.Context, authorizationToken string) error {
	return s.client.UpdateGitHubToken(ctx, s.sandboxID, authorizationToken)
}

// UpdateGitRepositoryResourceToken 更新当前沙箱中指定 Git 仓库资源的授权令牌。
// 具体行为参见 [Client.UpdateGitRepositoryResourceToken]。
func (s *Sandbox) UpdateGitRepositoryResourceToken(ctx context.Context, resourceID, authorizationToken string) error {
	return s.client.UpdateGitRepositoryResourceToken(ctx, s.sandboxID, resourceID, authorizationToken)
}

// WaitForReady 轮询当前沙箱的状态，直到进入 running 状态或上下文被取消。
// 具体行为参见 [Client.WaitForReady]。
func (s *Sandbox) WaitForReady(ctx context.Context, opts ...PollOption) (*SandboxInfo, error) {
	return s.client.WaitForReady(ctx, s.sandboxID, opts...)
}
