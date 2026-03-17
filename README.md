# Prism

HTTP/HTTPS 请求中继工具，支持透明转发、基于标识符的路由、TUI 管理、Header 注入、IP 白名单、API Key 认证和按用户限流。

## 功能特性

### 核心功能

- **三层路由结构**: Source -> Project -> RouteRule，灵活的组织架构
- **五种匹配模式**:
  - `param_path` - 参数化路径匹配 (`/users/{id}`)
  - `url_param` - URL参数匹配 (`?tenant=acme`)
  - `request_body` - JSON请求体匹配
  - `request_form` - 表单数据匹配
  - `plugin` - 自定义插件匹配
- **CEL表达式引擎**: 灵活的条件判断和参数提取
- **插件系统**: 基于进程隔离的自定义路由插件

### 其他功能

- **透明转发**: 类似 Nginx 反向代理，原封不动转发请求到目标
- **TUI 管理**: Bubble Tea 框架终端界面管理配置
- **Header 注入**: 支持标准和自定义 HTTP Header
- **IP 白名单**: 支持单 IP 和 CIDR 范围
- **API Key 认证**: 请求认证，提取用户身份
- **按用户限流**: 令牌桶算法，每用户独立配置
- **Prometheus 监控**: /metrics 端点暴露指标

## 平台支持

| 平台 | 架构 | 支持状态 |
|------|------|----------|
| Linux | amd64, arm64 | 完全支持 |
| macOS | amd64, arm64 | 完全支持 |
| FreeBSD | amd64, arm64 | 完全支持 |
| Windows | 所有 | **不支持** (Go插件限制) |

> **重要提示**: 插件必须与主程序使用相同的Go版本编译，且必须在同一平台上编译。

## 安装

```bash
# 构建
go build -o prism .

# 或使用 Makefile
make build
```

## 快速开始

### 1. 初始化数据库

```bash
./prism migrations up
```

### 2. 配置来源和项目

```sql
-- 添加来源
INSERT INTO source (id, name, description, enabled)
VALUES ('src-001', 'weixin', '微信公众号来源', 1);

-- 添加项目
INSERT INTO project (id, source_id, name, description, enabled)
VALUES ('proj-001', 'src-001', 'order-service', '订单服务', 1);
```

### 3. 配置路由规则

```sql
-- 参数化路径匹配
INSERT INTO route_rule (id, project_id, name, match_type, path_pattern, target_url, priority, enabled)
VALUES (
    'rule-001',
    'proj-001',
    'get-order',
    'param_path',
    '/orders/{orderId}',
    'http://backend:8080/api/orders/{orderId}',
    1,
    1
);

-- 带CEL表达式的规则
INSERT INTO route_rule (id, project_id, name, match_type, path_pattern, cel_expression, target_url, priority, enabled)
VALUES (
    'rule-002',
    'proj-001',
    'create-order',
    'param_path',
    '/orders',
    'method == ''POST'' && headers[''Content-Type''] == ''application/json''',
    'http://backend:8080/api/orders',
    2,
    1
);

-- URL参数匹配
INSERT INTO route_rule (id, project_id, name, match_type, cel_expression, target_url, priority, enabled)
VALUES (
    'rule-003',
    'proj-001',
    'query-orders',
    'url_param',
    'params.action == ''list''',
    'http://backend:8080/api/orders/list',
    3,
    1
);
```

### 4. 启动服务

```bash
# 使用默认配置
./prism run

# 指定配置文件
./prism run -c /path/to/config.toml

# 指定监听地址
./prism run -a :9090
```

### 5. 测试请求

```bash
# 路径参数匹配
curl http://localhost:8080/weixin/orders/12345

# URL参数匹配
curl "http://localhost:8080/weixin/orders?action=list"
```

## 使用方法

### 启动代理服务

```bash
# 使用默认配置
./prism run

# 指定配置文件
./prism run -c /path/to/config.toml

# 指定监听地址
./prism run -a :9090
```

### 启动 TUI 管理界面

```bash
./prism tui
```

### 路由管理

```bash
# 列出所有路由
./prism route list

# 添加路由
./prism route add

# 删除路由
./prism route delete <id>
```

### API Key 管理

```bash
# 列出所有 API Key
./prism apikey list

# 生成新 Key
./prism apikey generate

# 删除 Key
./prism apikey delete <id>
```

### 版本信息

```bash
./prism version
```

## 配置文件

配置文件使用 TOML 格式：

```toml
[sdk]
    [sdk.log]
        level = "debug"
        filename = "prism.log"
        [sdk.log.rotate]
            max_age = 168
            rotation_time = 24

    [sdk.sqlite]
        db = "prism.db"

[app]
    [app.server]
        host = "0.0.0.0"
        port = 8080
        tls_port = 8443

    [app.proxy]
        read_timeout = 30
        write_timeout = 30
        idle_timeout = 120

    [app.ratelimit]
        window_size = 10
        limit = 100

    # 插件配置
    [app.plugin]
        paths = ["./plugins", "/usr/local/lib/prism/plugins"]
```

## CEL 表达式

Prism 使用 CEL (Common Expression Language) 进行灵活的条件匹配和参数提取。

### 可用变量

| 变量 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `path` | map | 路径参数 | `path.tenant == 'acme'` |
| `params` | map | URL参数 | `params.action == 'create'` |
| `headers` | map | 请求头 | `headers['Content-Type'] == 'application/json'` |
| `body` | map | JSON请求体 | `body['user']['role'] == 'admin'` |
| `method` | string | HTTP方法 | `method == 'POST'` |
| `host` | string | 主机名 | `host == 'api.example.com'` |
| `pathStr` | string | 原始路径 | `pathStr.startsWith('/api/v1')` |

### 表达式示例

```javascript
// 简单条件
"path.tenant == 'acme'"

// 多条件组合
"path.tenant == 'acme' && params.action == 'create'"

// 请求头检查
"headers['Authorization'].startsWith('Bearer')"

// 请求体字段检查
"body['user']['role'] == 'admin' || body['user']['role'] == 'superadmin'"

// 复杂逻辑
"(path.tenant == 'acme' || path.tenant == 'beta') && method == 'POST'"
```

### 支持的函数

- 字符串: `startsWith()`, `endsWith()`, `contains()`
- 比较: `==`, `!=`, `<`, `>`, `<=`, `>=`
- 逻辑: `&&`, `||`, `!`
- 集合: `in`

## 插件开发

Prism 支持自定义路由插件，用于处理复杂的匹配逻辑。详见 [插件开发指南](docs/PLUGIN_DEVELOPMENT.md)。

### 快速示例

```go
package main

import (
    "context"
    "github.com/rfancn/prism/plugin"
)

type MyPlugin struct{}

func (p *MyPlugin) Info(ctx context.Context) (*plugin.PluginInfo, error) {
    return &plugin.PluginInfo{
        Name:        "my-plugin",
        Description: "我的自定义路由插件",
        Version:     "1.0.0",
    }, nil
}

func (p *MyPlugin) Match(ctx context.Context, req *plugin.MatchRequest) (*plugin.MatchResponse, error) {
    // 实现自定义匹配逻辑
    if req.Headers["X-Special-Header"] == "secret-value" {
        return &plugin.MatchResponse{
            Matched: true,
            Params:  map[string]string{"type": "special"},
        }, nil
    }
    return &plugin.MatchResponse{Matched: false}, nil
}

func main() {
    plugin.Serve(&MyPlugin{})
}
```

## API 端点

| 端点 | 说明 |
|------|------|
| `GET /health` | 健康检查 |
| `GET /ready` | 就绪检查 |
| `GET /metrics` | Prometheus 指标 |
| `/{source}/...` | 路由转发（根据配置匹配） |

## 监控指标

| 指标 | 说明 |
|------|------|
| `prism_requests_total` | 总请求数 |
| `prism_request_duration_seconds` | 请求延迟 |
| `prism_active_connections` | 活跃连接数 |
| `prism_proxy_requests_total` | 代理请求数 |
| `prism_rate_limit_hits_total` | 限流触发次数 |
| `prism_auth_failures_total` | 认证失败次数 |

## 开发

```bash
# 生成 sqlc 代码
go generate ./...

# 运行测试
go test ./...

# 代码检查
go vet ./...
```

## 许可证

MIT License