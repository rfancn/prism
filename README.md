# Prism

HTTP/HTTPS 请求中继工具，支持透明转发、基于标识符的路由、TUI 管理、Header 注入、IP 白名单、API Key 认证和按用户限流。

## 功能特性

- **透明转发**：类似 Nginx 反向代理，原封不动转发请求到目标
- **标识符路由**：支持 Path、JSON Body、URL 参数三种解析方式
- **TUI 管理**：Bubble Tea 框架终端界面管理配置
- **Header 注入**：支持标准和自定义 HTTP Header
- **IP 白名单**：支持单 IP 和 CIDR 范围
- **API Key 认证**：请求认证，提取用户身份
- **按用户限流**：令牌桶算法，每用户独立配置
- **Prometheus 监控**：/metrics 端点暴露指标

## 安装

```bash
# 构建
go build -o prism .

# 或使用 Makefile
make build
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
```

## API 端点

| 端点 | 说明 |
|------|------|
| `GET /health` | 健康检查 |
| `GET /ready` | 就绪检查 |
| `GET /metrics` | Prometheus 指标 |
| `/api/*` | 保护路由（需要认证） |
| `/public/*` | 公开路由 |

## 标识符解析

支持三种标识符解析方式：

### 1. Path Parameter

```
Pattern: /api/{tenant}/users
Request: /api/acme/users
Extract: tenant = "acme"
```

### 2. JSON Body

```json
{
  "user": {
    "tenant_id": "acme"
  }
}
```

配置字段名：`user.tenant_id`（支持嵌套）

### 3. URL Parameter

```
Request: /api/users?tenant_id=acme
Extract: tenant_id = "acme"
```

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