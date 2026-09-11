# Go SDK 与 e2b JS SDK API 对照

本文用于说明七牛云 Go SDK 与 e2b JS SDK 的沙箱 API 关系，帮助跨语言迁移和维护两端的一致产品语义。

Go SDK 的沙箱控制面 API 设计时尽量参考 e2b JS SDK 的方法语义和参数顺序，同时遵循 Go 的调用习惯：显式传递 `context.Context`，通过返回值返回错误，并使用结构体和选项函数表达可选参数。

## 1. API 对象模型

两端的主要差异是调用入口：

| 使用场景 | e2b JS SDK | Go SDK |
| --- | --- | --- |
| 创建、连接或列出沙箱 | `Sandbox.create()`、`Sandbox.connect()`、`Sandbox.list()` | `Client.Create()`、`Client.Connect()`、`Client.List()` |
| 操作已有沙箱 | `sandbox.getInfo()`、`sandbox.kill()` | `sb.GetInfo()`、`sb.Kill()` |
| 只持有沙箱 ID | 通过静态 API 或重新连接 | `client.GetInfo(ctx, sandboxID)`、`client.Kill(ctx, sandboxID)` |
| 操作沙箱内部 envd | `sandbox.files`、`sandbox.commands` | `sb.Files()`、`sb.Commands()` |

Go SDK 同时提供 `Client` 和 `Sandbox` 两套控制面入口：

- `Client` 方法显式接收 `sandboxID`，适合从列表结果或持久化数据中操作沙箱。
- `Sandbox` 方法自动使用实例自身的 ID，适合已经持有沙箱实例的场景。
- 两套方法使用同一套控制面实现，行为和错误处理保持一致。

## 2. 已有对应关系的 API

### 2.1 创建、连接和列出沙箱

| e2b JS SDK | Go SDK |
| --- | --- |
| `Sandbox.create()` | `client.Create(ctx, params)` |
| `Sandbox.connect(sandboxId)` | `client.Connect(ctx, sandboxID, params)` |
| `Sandbox.list()` | `client.List(ctx, params)` |

### 2.2 生命周期和状态查询

| e2b JS SDK | Go SDK：`Client` | Go SDK：`Sandbox` |
| --- | --- | --- |
| `sandbox.getInfo()` | `GetInfo(ctx, sandboxID)` | `sb.GetInfo(ctx)` |
| `sandbox.getMetrics()` | `GetMetrics(ctx, sandboxID, params)` | `sb.GetMetrics(ctx, params)` |
| `sandbox.kill()` | `Kill(ctx, sandboxID)` | `sb.Kill(ctx)` |
| `sandbox.pause()` | `Pause(ctx, sandboxID)` | `sb.Pause(ctx)` |
| `sandbox.setTimeout(timeout)` | `SetTimeout(ctx, sandboxID, timeout)` | `sb.SetTimeout(ctx, timeout)` |
| `sandbox.isRunning()` | — | `sb.IsRunning(ctx)` |

### 2.3 envd 能力

| e2b JS SDK | Go SDK |
| --- | --- |
| `sandbox.files` | `sb.Files()` |
| `sandbox.commands` | `sb.Commands()` |
| `sandbox.pty` | `sb.Pty()` |
| `sandbox.git` | `sb.Git()` |
| `sandbox.getHost(port)` | `sb.GetHost(port)` |
| `sandbox.uploadUrl(path)` | `sb.UploadURL(path, opts...)` |
| `sandbox.downloadUrl(path)` | `sb.DownloadURL(path, opts...)` |

这些能力依赖具体沙箱实例和 envd 连接，因此不放在 `Client` 上。

### 2.4 GitHub 和资源令牌

| HTTP API | Go SDK：`Client` | Go SDK：`Sandbox` |
| --- | --- | --- |
| `PUT /sandboxes/{sandboxID}/github-token` | `UpdateGitHubToken(ctx, sandboxID, token)` | `sb.UpdateGitHubToken(ctx, token)` |
| `PATCH /sandboxes/{sandboxID}/resources/{resourceID}` | `UpdateGitRepositoryResourceToken(ctx, sandboxID, resourceID, token)` | `sb.UpdateGitRepositoryResourceToken(ctx, resourceID, token)` |
| `GET /sandboxes/{sandboxID}/resources` | `GetResources(ctx, sandboxID)` | `sb.GetResources(ctx)` |

资源级 PATCH 接口只更新指定 GitHub 资源的授权令牌；沙箱级 GitHub Token 接口更新指定沙箱使用的 GitHub 授权令牌，两者的作用范围不同。

## 3. Go SDK 的产品扩展 API

以下 Go API 在当前 e2b JS SDK 公共源码中没有同名的一对一方法，属于 Go SDK 当前产品能力：

| Go SDK | 作用 |
| --- | --- |
| `GetLogs` | 查询沙箱日志 |
| `Refresh` | 延长沙箱存活时间 |
| `WaitForReady` | 轮询等待沙箱进入 `running` 状态 |
| `GetInjections` / `UpdateInjections` | 查询和更新运行时请求注入规则 |
| `GetResources` | 查询沙箱已挂载的资源 |
| `UpdateGitHubToken` | 更新沙箱级 GitHub 授权令牌 |
| `UpdateGitRepositoryResourceToken` | 更新指定 GitHub 资源的授权令牌 |

其中 `Refresh` 在 e2b JS SDK 中没有同名方法；需要调整存活时间时通常使用 `setTimeout`。

## 4. E2B JS SDK 中 Go SDK 尚未提供的 API

以下接口存在于当前 e2b JS SDK，但当前 Go SDK 尚未暴露对应的公开 API：

| e2b JS SDK | 作用 | Go SDK 状态 |
| --- | --- | --- |
| `Sandbox.fork(sandboxId, opts)` / `sandbox.fork(opts)` | 从运行中的沙箱创建一个或多个副本 | 暂无对应 API |
| `sandbox.updateNetwork(network, opts)` | 更新运行中沙箱的出站网络配置 | 暂无对应 API；Go 目前仅支持创建沙箱时通过 `CreateParams.Network` 配置 |
| `sandbox.createSnapshot(opts)` | 将当前沙箱状态创建为持久化快照 | 暂无对应 API |
| `Sandbox.listSnapshots(opts)` | 分页列出全部快照 | 暂无对应 API |
| `sandbox.listSnapshots(opts)` | 分页列出当前沙箱创建的快照 | 暂无对应 API |
| `Sandbox.deleteSnapshot(snapshotId, opts)` | 删除快照 | 暂无对应 API |
| `sandbox.getMcpUrl()` | 获取沙箱内 MCP Gateway 地址 | 暂无对应 API |
| `sandbox.getMcpToken()` | 获取沙箱内 MCP Gateway Token | 暂无对应 API |

### 4.1 快照 API 关系

快照是可持久化、可复用的沙箱状态，不是运行中的沙箱列表：

1. `createSnapshot` 创建快照并返回 `snapshotId`。
2. `listSnapshots` 通过分页查询快照，可按来源沙箱、名称和分页游标过滤。
3. `snapshotId` 可以再次作为 `Sandbox.create(snapshotId)` 的模板参数。

e2b JS SDK 当前使用以下控制面请求：

- 创建快照：`POST /sandboxes/{sandboxID}/snapshots`
- 查询快照：`GET /snapshots`，支持 `sandboxID`、`name`、`limit` 和 `nextToken`
- 删除快照：`DELETE /templates/{templateID}`

`Sandbox.list()` 列出运行中或已暂停的沙箱；`listSnapshots()` 列出可复用的持久化快照。

## 5. Go 与 JS 的调用差异

### 5.1 上下文和错误

Go API 将请求上下文放在第一个参数，并通过 `error` 返回网络错误、服务端错误和参数错误：

```go
info, err := client.GetInfo(ctx, sandboxID)
if err != nil {
	return fmt.Errorf("get sandbox info: %w", err)
}
```

JS SDK 通常通过 `Promise` 抛出异常；迁移时需要将 `await` 调用改为 Go 的返回值和错误判断。

### 5.2 可选参数

JS SDK 常使用对象选项，例如 `getMetrics({ start, end })`。Go SDK 使用参数结构体表达请求参数，使用选项函数表达轮询行为：

```go
start := time.Now().Add(-time.Hour).Unix()
end := time.Now().Unix()

metrics, err := client.GetMetrics(ctx, sandboxID, &sandbox.GetMetricsParams{
	Start: &start,
	End:   &end,
})

info, err := client.WaitForReady(ctx, sandboxID,
	sandbox.WithPollInterval(time.Second),
)
```

### 5.3 时间类型

Go SDK 使用 `time.Duration` 表示超时时间和轮询间隔；JS SDK 通常使用毫秒数。迁移时应明确换算：

```go
timeout := 30 * time.Second
```

## 6. 参考

- [e2b JS SDK Sandbox API](https://github.com/e2b-dev/E2B/blob/main/packages/js-sdk/src/sandbox/index.ts)
- [e2b JS SDK Sandbox API 类型定义](https://github.com/e2b-dev/E2B/blob/main/packages/js-sdk/src/sandbox/sandboxApi.ts)
- [Go SDK sandbox 包文档](../../sandbox/doc.go)

本文以当前 Go SDK API 为准；e2b JS SDK 对照以目标版本源码为准。新增能力或两端 API 发生变化时，应同步更新本文档。
