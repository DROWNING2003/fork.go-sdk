package sandbox

import (
	"fmt"

	"github.com/qiniu/go-sdk/v7/sandbox/internal/apis"
)

// ---------------------------------------------------------------------------
// SDK 自有类型 — 沙箱资源
// ---------------------------------------------------------------------------

// GitRepositoryType Git 仓库托管平台类型。
type GitRepositoryType string

// Git 仓库托管平台类型常量。后续若服务端扩展更多类型，将在此新增对应常量。
const (
	// GitRepositoryTypeGithub GitHub 仓库。
	GitRepositoryTypeGithub GitRepositoryType = "github_repository"
)

// GitRepositoryResource Git 仓库资源，沙箱启动前由平台拉取并挂载快照到指定路径。
type GitRepositoryResource struct {
	// Type 仓库托管平台类型（必填）。当前仅支持 GitRepositoryTypeGithub。
	Type GitRepositoryType

	// URL 仓库 URL（HTTPS 或 SSH 形式），如
	// https://github.com/owner/repo.git 或 git@github.com:owner/repo.git。
	URL string

	// MountPath 仓库内容在沙箱内的绝对挂载路径。
	MountPath string

	// AuthorizationToken 用于克隆该仓库的访问 token。
	AuthorizationToken *string
}

// KodoResource Kodo 存储桶资源，沙箱启动前由平台通过 NFS 代理挂载到指定路径。
// 可通过 AccessKey 和 SecretKey 指定访问凭证；省略时使用 Client 配置的 Qiniu AK/SK 凭证。
type KodoResource struct {
	// AccessKey Kodo 访问密钥；与 SecretKey 必须同时指定。
	AccessKey *string

	// Bucket Kodo 存储桶名称（必填）。
	Bucket string

	// MountPath 存储桶内容在沙箱内的绝对挂载路径（必填）。
	MountPath string

	// Prefix 存储桶内可选的对象名前缀；不设置时挂载整个存储桶根目录。
	Prefix *string

	// ReadOnly 是否以只读方式挂载；当 AK/SK 缺少写权限时服务端也会自动只读。
	ReadOnly *bool

	// SecretKey Kodo 私钥；与 AccessKey 必须同时指定。
	SecretKey *string
}

// SandboxResourceSpec 沙箱资源规约（discriminated union），各字段互斥，只能设置一个。
type SandboxResourceSpec struct {
	// GitRepository GitHub 仓库资源。
	GitRepository *GitRepositoryResource

	// Kodo Kodo 存储桶资源。
	Kodo *KodoResource
}

// GitRepositoryResourceInfo 是查询到的 Git 仓库资源，不包含授权令牌。
type GitRepositoryResourceInfo struct {
	// ResourceID 服务端生成的资源 ID。
	ResourceID string

	// Type 仓库托管平台类型。
	Type GitRepositoryType

	// URL 仓库 URL。
	URL string

	// MountPath 仓库内容在沙箱内的绝对挂载路径。
	MountPath string
}

// KodoResourceInfo 是查询到的 Kodo 存储桶资源，不包含访问凭证。
type KodoResourceInfo struct {
	// ResourceID 服务端生成的资源 ID。
	ResourceID string

	// Bucket Kodo 存储桶名称。
	Bucket string

	// MountPath 存储桶内容在沙箱内的绝对挂载路径。
	MountPath string

	// Prefix 存储桶内可选的对象名前缀。
	Prefix *string

	// ReadOnly 是否以只读方式挂载。
	ReadOnly *bool
}

// SandboxResourceInfo 是查询到的沙箱资源。凭证字段不会出现在返回值中。
type SandboxResourceInfo struct {
	// GitRepository Git 仓库资源。
	GitRepository *GitRepositoryResourceInfo

	// Kodo Kodo 存储桶资源。
	Kodo *KodoResourceInfo
}

// ---------------------------------------------------------------------------
// 转换函数 — SDK → apis
// ---------------------------------------------------------------------------

func sandboxResourceSpecToAPI(spec SandboxResourceSpec) (apis.SandboxResource, error) {
	var r apis.SandboxResource
	count := 0
	if spec.GitRepository != nil {
		count++
	}
	if spec.Kodo != nil {
		count++
	}
	if count == 0 {
		return r, fmt.Errorf("SandboxResourceSpec: exactly one resource type must be set (GitRepository or Kodo), got none")
	}
	if count > 1 {
		return r, fmt.Errorf("SandboxResourceSpec: exactly one resource type must be set, but got %d", count)
	}

	switch {
	case spec.GitRepository != nil:
		if spec.GitRepository.Type == "" {
			return r, fmt.Errorf("GitRepositoryResource.Type must be set (e.g. GitRepositoryTypeGithub)")
		}
		if spec.GitRepository.URL == "" {
			return r, fmt.Errorf("GitRepositoryResource.URL must be set")
		}
		if spec.GitRepository.MountPath == "" {
			return r, fmt.Errorf("GitRepositoryResource.MountPath must be set")
		}
		if spec.GitRepository.AuthorizationToken == nil || *spec.GitRepository.AuthorizationToken == "" {
			return r, fmt.Errorf("GitRepositoryResource.AuthorizationToken must be set")
		}
		if err := r.FromGitRepositoryResource(apis.GitRepositoryResource{
			URL:                spec.GitRepository.URL,
			MountPath:          spec.GitRepository.MountPath,
			AuthorizationToken: spec.GitRepository.AuthorizationToken,
			Type:               apis.GitRepositoryResourceType(spec.GitRepository.Type),
		}); err != nil {
			return r, err
		}
	case spec.Kodo != nil:
		if spec.Kodo.Bucket == "" {
			return r, fmt.Errorf("KodoResource.Bucket must be set")
		}
		if spec.Kodo.MountPath == "" {
			return r, fmt.Errorf("KodoResource.MountPath must be set")
		}
		if (spec.Kodo.AccessKey == nil) != (spec.Kodo.SecretKey == nil) {
			return r, fmt.Errorf("KodoResource.AccessKey and SecretKey must be set together")
		}
		if spec.Kodo.AccessKey != nil && (*spec.Kodo.AccessKey == "" || *spec.Kodo.SecretKey == "") {
			return r, fmt.Errorf("KodoResource.AccessKey and SecretKey must not be empty")
		}
		if err := r.FromKodoResource(apis.KodoResource{
			AccessKey: spec.Kodo.AccessKey,
			Bucket:    spec.Kodo.Bucket,
			MountPath: spec.Kodo.MountPath,
			Prefix:    spec.Kodo.Prefix,
			ReadOnly:  spec.Kodo.ReadOnly,
			SecretKey: spec.Kodo.SecretKey,
		}); err != nil {
			return r, err
		}
	default:
		return r, fmt.Errorf("SandboxResourceSpec: unsupported resource type")
	}
	return r, nil
}

func sandboxResourceInfoFromAPI(resource apis.SandboxResource) (SandboxResourceInfo, error) {
	discriminator, err := resource.Discriminator()
	if err != nil {
		return SandboxResourceInfo{}, err
	}
	switch discriminator {
	case string(apis.GithubRepository):
		value, err := resource.AsGitRepositoryResource()
		if err != nil {
			return SandboxResourceInfo{}, err
		}
		return SandboxResourceInfo{GitRepository: &GitRepositoryResourceInfo{
			ResourceID: derefString(value.ResourceID),
			Type:       GitRepositoryType(value.Type),
			URL:        value.URL,
			MountPath:  value.MountPath,
		}}, nil
	case string(apis.Kodo):
		value, err := resource.AsKodoResource()
		if err != nil {
			return SandboxResourceInfo{}, err
		}
		return SandboxResourceInfo{Kodo: &KodoResourceInfo{
			ResourceID: derefString(value.ResourceID),
			Bucket:     value.Bucket,
			MountPath:  value.MountPath,
			Prefix:     value.Prefix,
			ReadOnly:   value.ReadOnly,
		}}, nil
	default:
		return SandboxResourceInfo{}, fmt.Errorf("unknown sandbox resource type: %s", discriminator)
	}
}

func sandboxResourceInfosFromAPI(resources []apis.SandboxResource) ([]SandboxResourceInfo, error) {
	if resources == nil {
		return nil, nil
	}
	result := make([]SandboxResourceInfo, len(resources))
	for i, resource := range resources {
		info, err := sandboxResourceInfoFromAPI(resource)
		if err != nil {
			return nil, err
		}
		result[i] = info
	}
	return result, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
