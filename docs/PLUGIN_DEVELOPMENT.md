# Prism 插件开发指南

本文档介绍如何为 Prism 开发自定义路由插件。

## 概述

Prism 采用基于 [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin) 的插件系统，支持通过 gRPC 协议与插件通信。插件作为独立进程运行，与主程序隔离，确保稳定性和安全性。

## 平台支持

### 支持的平台

| 平台 | 架构 | 支持状态 |
|------|------|----------|
| Linux | amd64, arm64 | 完全支持 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) | 完全支持 |
| FreeBSD | amd64, arm64 | 完全支持 |

### Go 版本兼容性

**关键要求**: 插件必须与 Prism 主程序使用完全相同的 Go 版本编译。

```bash
# 检查 Prism 编译使用的 Go 版本
./prism version

# 使用相同版本编译插件
go build -o my-plugin .
```

**注意事项:**
- Go 1.x 的不同小版本（如 1.21.0 和 1.21.1）之间插件可能不兼容
- 插件必须在与 Prism 运行环境相同的操作系统和架构上编译
- 跨平台编译的插件无法加载

### 编译注意事项

```bash
# 正确的编译方式（与主程序相同环境）
go build -o my-plugin .

# 错误示例 - 不要这样做
GOOS=linux GOARCH=amd64 go build -o my-plugin .  # 交叉编译会导致插件无法加载
```

### 部署检查清单

在部署插件前，请确认：

- [ ] 插件与 Prism 运行在相同的操作系统上
- [ ] 插件与 Prism 使用相同的 Go 版本编译
- [ ] 插件文件具有执行权限 (`chmod +x my-plugin`)
- [ ] 插件路径在 Prism 配置的插件目录中

## 插件接口

### RouterPlugin 接口

所有路由插件必须实现 `RouterPlugin` 接口：

```go
type RouterPlugin interface {
    // Info 返回插件信息
    Info(ctx context.Context) (*PluginInfo, error)

    // Match 执行路由匹配
    Match(ctx context.Context, req *MatchRequest) (*MatchResponse, error)
}
```

### 数据结构

#### MatchRequest

```go
type MatchRequest struct {
    Method     string            // HTTP方法 (GET, POST, PUT, DELETE等)
    Headers    map[string]string // HTTP请求头
    URLParams  map[string]string // URL查询参数
    PathParams map[string]string // 路径参数
    Body       []byte            // 请求体
    RemoteAddr string            // 客户端地址
    Path       string            // 请求路径
}
```

#### MatchResponse

```go
type MatchResponse struct {
    Matched bool              // 是否匹配成功
    Params  map[string]string // 提取的参数
    Error   string            // 错误信息
    RouteID string            // 匹配的路由ID
}
```

#### PluginInfo

```go
type PluginInfo struct {
    Name        string // 插件名称
    Description string // 插件描述
    Version     string // 插件版本
    Author      string // 作者
}
```

## 快速开始

### 1. 创建插件项目

```bash
mkdir my-router-plugin
cd my-router-plugin
go mod init my-router-plugin
```

### 2. 添加依赖

```bash
go get github.com/rfancn/prism/plugin
go get github.com/hashicorp/go-plugin
```

### 3. 实现插件

创建 `main.go`：

```go
package main

import (
    "context"
    "strings"

    "github.com/rfancn/prism/plugin"
)

// MyPlugin 自定义插件实现
type MyPlugin struct{}

// Info 返回插件信息
func (p *MyPlugin) Info(ctx context.Context) (*plugin.PluginInfo, error) {
    return &plugin.PluginInfo{
        Name:        "my-plugin",
        Description: "我的自定义路由插件",
        Version:     "1.0.0",
        Author:      "Your Name",
    }, nil
}

// Match 执行路由匹配
func (p *MyPlugin) Match(ctx context.Context, req *plugin.MatchRequest) (*plugin.MatchResponse, error) {
    // 实现你的匹配逻辑
    if strings.HasPrefix(req.Path, "/api/v1/") {
        return &plugin.MatchResponse{
            Matched: true,
            Params: map[string]string{
                "version": "v1",
            },
        }, nil
    }

    return &plugin.MatchResponse{
        Matched: false,
    }, nil
}

func main() {
    // 启动插件服务
    plugin.Serve(&MyPlugin{})
}
```

### 4. 编译插件

```bash
go build -o my-plugin .
```

### 5. 部署插件

将编译后的二进制文件放到 Prism 的插件目录中：

```bash
cp my-plugin /path/to/prism/plugins/
```

## 匹配逻辑示例

### 基于路径匹配

```go
func (p *MyPlugin) Match(ctx context.Context, req *plugin.MatchRequest) (*plugin.MatchResponse, error) {
    // 匹配 /api/{tenant}/users 路径
    parts := strings.Split(req.Path, "/")
    if len(parts) >= 4 && parts[1] == "api" && parts[3] == "users" {
        return &plugin.MatchResponse{
            Matched: true,
            Params: map[string]string{
                "tenant": parts[2],
            },
        }, nil
    }
    return &plugin.MatchResponse{Matched: false}, nil
}
```

### 基于请求头匹配

```go
func (p *MyPlugin) Match(ctx context.Context, req *plugin.MatchRequest) (*plugin.MatchResponse, error) {
    // 检查 X-Tenant-ID 请求头
    tenantID := req.Headers["X-Tenant-ID"]
    if tenantID != "" {
        return &plugin.MatchResponse{
            Matched: true,
            Params: map[string]string{
                "tenant_id": tenantID,
            },
        }, nil
    }
    return &plugin.MatchResponse{Matched: false}, nil
}
```

### 基于查询参数匹配

```go
func (p *MyPlugin) Match(ctx context.Context, req *plugin.MatchRequest) (*plugin.MatchResponse, error) {
    // 检查 tenant 查询参数
    tenant := req.URLParams["tenant"]
    if tenant != "" {
        return &plugin.MatchResponse{
            Matched: true,
            Params: map[string]string{
                "tenant": tenant,
            },
        }, nil
    }
    return &plugin.MatchResponse{Matched: false}, nil
}
```

### 基于请求体匹配

```go
import "encoding/json"

func (p *MyPlugin) Match(ctx context.Context, req *plugin.MatchRequest) (*plugin.MatchResponse, error) {
    // 只处理 JSON 请求体
    contentType := req.Headers["Content-Type"]
    if contentType != "application/json" {
        return &plugin.MatchResponse{Matched: false}, nil
    }

    // 解析 JSON
    var body struct {
        TenantID string `json:"tenant_id"`
    }
    if err := json.Unmarshal(req.Body, &body); err != nil {
        return &plugin.MatchResponse{
            Matched: false,
            Error:   "无效的JSON格式",
        }, nil
    }

    if body.TenantID != "" {
        return &plugin.MatchResponse{
            Matched: true,
            Params: map[string]string{
                "tenant_id": body.TenantID,
            },
        }, nil
    }

    return &plugin.MatchResponse{Matched: false}, nil
}
```

## 最佳实践

### 1. 错误处理

- 在 `Match` 方法中返回明确的错误信息
- 使用 `Error` 字段传递用户友好的错误消息
- 不要 panic，优雅处理所有错误情况

### 2. 性能优化

- 避免在 `Match` 方法中进行耗时操作
- 缓存可以复用的计算结果
- 使用高效的正则表达式或字符串操作

### 3. 日志记录

- 使用结构化日志
- 记录关键操作和错误
- 避免敏感信息泄露

### 4. 测试

为你的插件编写测试：

```go
package main

import (
    "context"
    "testing"

    "github.com/rfancn/prism/plugin"
)

func TestMyPlugin_Match(t *testing.T) {
    p := &MyPlugin{}

    tests := []struct {
        name     string
        req      *plugin.MatchRequest
        expected bool
    }{
        {
            name: "匹配成功",
            req: &plugin.MatchRequest{
                Path: "/api/v1/users",
            },
            expected: true,
        },
        {
            name: "匹配失败",
            req: &plugin.MatchRequest{
                Path: "/api/v2/users",
            },
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := p.Match(context.Background(), tt.req)
            if err != nil {
                t.Fatalf("Match returned error: %v", err)
            }
            if resp.Matched != tt.expected {
                t.Errorf("Expected matched=%v, got %v", tt.expected, resp.Matched)
            }
        })
    }
}
```

## 插件生命周期

1. **加载**: Prism 启动时扫描插件目录，加载所有插件
2. **初始化**: 调用 `Info` 方法获取插件信息
3. **运行**: 处理请求时调用 `Match` 方法
4. **卸载**: Prism 关闭时终止所有插件进程

## 调试插件

### 启用日志

设置环境变量启用插件日志：

```bash
export PLUGIN_LOG_LEVEL=debug
```

### 独立测试

可以直接运行插件进程进行测试：

```bash
# 设置插件模式
export PLUGIN_MODE=standalone

# 运行插件
./my-plugin
```

## 常见问题

### Q: 插件加载失败？

检查：
- 插件文件是否有执行权限 (`chmod +x my-plugin`)
- 握手配置是否正确
- gRPC 端口是否被占用
- **Go 版本是否与主程序一致**
- **操作系统是否支持**（Windows 不支持）

### Q: 插件版本不兼容错误？

错误信息示例：
```
plugin was built with a different version of package X
```

解决方案：
1. 确保插件与 Prism 主程序使用完全相同的 Go 版本
2. 确保所有依赖版本一致
3. 在相同的环境中重新编译插件

### Q: Windows 上无法加载插件？

这是 Go 语言插件系统的限制。解决方案：
- 在 Linux/macOS/FreeBSD 上部署 Prism 和插件
- 考虑使用 Docker 容器运行 Prism

### Q: 插件匹配耗时太长？

建议：
- 优化匹配算法
- 减少不必要的 I/O 操作
- 考虑使用缓存

### Q: 如何处理并发请求？

插件系统会为每个请求创建独立的 goroutine，确保：
- 避免使用全局可变状态
- 使用线程安全的数据结构
- 必要时使用 sync 包进行同步

## 示例项目

参考 `plugins/example/` 目录中的示例插件，了解完整实现。