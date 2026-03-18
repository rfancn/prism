# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Prism is an HTTP/HTTPS request relay tool written in Go. It supports transparent forwarding, identifier-based routing, TUI management, header injection, IP whitelisting, API key authentication, and per-user rate limiting.

**Key Features:**
- **Three-layer routing structure**: Source -> Project -> RouteRule
- **Five matching modes**: param_path, url_param, request_body, request_form, plugin
- **CEL expression engine**: Flexible condition matching and parameter extraction
- **Plugin system**: Process-isolated plugins based on hashicorp/go-plugin

## Build & Development Commands

```bash
# Build the binary with version info
go build -o prism .
# or
make build

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a single test (e.g., specific package or function)
go test ./pkg/cel/... -v
go test -run TestEngineEvaluate ./pkg/cel/...

# Generate sqlc code (after modifying SQL schema/queries)
go generate ./...

# Run linter
go vet ./...

# Run the service (use default database path: prism.db)
./prism run

# Run the service with custom database path
./prism run --db /path/to/custom.db

# Run with custom host/port (overrides database config)
./prism run --host 0.0.0.0 --port 8080

# Run TUI management interface
./prism tui

# Run TUI with custom database path
./prism tui --db /path/to/custom.db
```

**Note**: 在国内网络环境下，需要设置 Go 代理：
```bash
export GOPROXY=https://goproxy.cn,direct
```

## Key Dependencies

- **SDK**: `github.com/hdget/sdk v0.5.2` - provides Logger and DB access via `sdk.Logger()` and `sdk.Db()`
- **HTTP Framework**: `github.com/gin-gonic/gin` - HTTP server and routing
- **Database**: SQLite with sqlc code generation (`github.com/hdget/sdk/providers/db/sqlite3/sqlc v0.0.3`)
- **CLI**: `github.com/spf13/cobra` - command-line interface
- **TUI**: `github.com/charmbracelet/bubbletea` - terminal UI framework
- **Monitoring**: `github.com/prometheus/client_golang` - metrics exposition
- **Rate Limiting**: `golang.org/x/time/rate` - token bucket algorithm
- **CEL Engine**: `github.com/google/cel-go` - Common Expression Language
- **Plugin System**: `github.com/hashicorp/go-plugin` - process-isolated plugins

## Architecture

### Request Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Request Flow                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  Client Request                                                             │
│       │                                                                     │
│       ▼                                                                     │
│  ┌─────────────┐                                                            │
│  │ Gin Server  │ ─── Extract source name from path: /{source}/...           │
│  └──────┬──────┘                                                            │
│         │                                                                   │
│         ▼                                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    Three-Layer Routing                               │   │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐     │   │
│  │  │   Source    │ -> │   Project   │ -> │     RouteRule       │     │   │
│  │  │  (来源层)   │    │  (项目层)   │    │    (路由规则层)      │     │   │
│  │  └─────────────┘    └─────────────┘    └─────────────────────┘     │   │
│  │                                                 │                   │   │
│  │                                                 ▼                   │   │
│  │                                    ┌─────────────────────────┐     │   │
│  │                                    │      Matcher Factory    │     │   │
│  │                                    │  ┌─────┬─────┬─────┬───┐│     │   │
│  │                                    │  │Path │URL  │Body │...││     │   │
│  │                                    │  └─────┴─────┴─────┴───┘│     │   │
│  │                                    └─────────────────────────┘     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                   │
│         ▼                                                                   │
│  ┌─────────────────┐                                                        │
│  │  CEL Expression │ ─── Evaluate conditions, extract params               │
│  │     Engine      │                                                        │
│  └─────────────────┘                                                        │
│         │                                                                   │
│         ▼                                                                   │
│  ┌─────────────────┐                                                        │
│  │  Proxy Handler  │ ─── Inject headers, forward to target                 │
│  └─────────────────┘                                                        │
│         │                                                                   │
│         ▼                                                                   │
│     Target Server                                                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Three-Layer Routing Structure

```
Source (来源)
├── name: "weixin"           # 来源名称，从请求路径第一段提取
├── enabled: true
│
└── Project (项目)
    ├── name: "order-service"    # 项目名称
    ├── enabled: true
    │
    └── RouteRule (路由规则)
        ├── name: "create-order"        # 规则名称
        ├── match_type: "param_path"    # 匹配类型
        ├── path_pattern: "/orders/{id}" # 路径模式
        ├── cel_expression: "path.id != ''" # CEL表达式
        ├── target_url: "http://backend/api/orders" # 目标URL
        ├── priority: 1                  # 优先级（数字越小越优先）
        └── Headers []                   # 注入的请求头
```

### Matcher Types

| Type | Description | Use Case |
|------|-------------|----------|
| `param_path` | Parameterized path matching | `/users/{id}/orders/{orderId}` |
| `url_param` | URL query parameter matching | `?tenant=acme&action=create` |
| `request_body` | JSON body field matching | `{"tenant_id": "acme"}` |
| `request_form` | Form data matching | `tenant=acme&action=create` |
| `plugin` | Custom plugin matching | Complex custom logic |

### Directory Structure

- `cmd/` - Cobra CLI commands (`run`, `tui`, `route`, `version`)
- `g/` - Global configuration structures (`g.Config`) and constants
- `pkg/` - Core packages:
  - `config/` - Configuration management (load from database)
  - `router/` - Three-layer routing, matching, and forwarding
  - `matcher/` - Request matcher implementations (5 types)
  - `cel/` - CEL expression engine with sandbox security
  - `middleware/` - Gin middleware (registration pattern with `Register()`/`Get()`)
  - `proxy/` - Reverse proxy with director-based request rewriting
  - `server/` - Gin HTTP server setup and route registration
  - `monitor/` - Prometheus metrics and health endpoints
  - `types/` - Shared types across packages (TLS config)
- `tui/` - Bubble Tea TUI application
  - `app.go` - Main TUI application entry point
  - `list.go`, `choice.go`, `form.go` - Reusable UI components
  - `model_*.go` - Feature-specific models (sources, projects, route_rules, whitelist)
- `repository/` - Data access layer using `repository.New()` for `db.Queries`
- `plugin/` - Plugin system implementation
  - `interface.go` - RouterPlugin interface definition
  - `manager.go` - Plugin lifecycle management
  - `client.go` / `server.go` - gRPC communication
- `assets/` - SQL schema (`schema/`) and queries (`queries/`) for sqlc; migrations (`migrations/`)
- `autogen/db/` - sqlc-generated code (do not edit manually)
- `docs/` - Documentation files

### Configuration Structure

**配置已完全迁移到数据库**，不再使用配置文件。

**数据库路径**：通过命令行参数 `--db` 指定，默认为 `prism.db`

**配置优先级**：命令行参数 > 数据库配置 > 默认值

**配置表 `app_config`**：
| Key | 默认值 | 说明 |
|-----|--------|------|
| `server.host` | `0.0.0.0` | 服务监听主机地址 |
| `server.port` | `8080` | HTTP 服务端口 |
| `server.tls_port` | `8443` | HTTPS 服务端口 |
| `proxy.read_timeout` | `30` | 代理读超时（秒）|
| `proxy.write_timeout` | `30` | 代理写超时（秒）|
| `proxy.idle_timeout` | `120` | 代理空闲超时（秒）|

Access via `g.Config` global variable after loading. Config structs use `mapstructure` tags.

### Database Schema

Key tables (see `assets/schema/schema.sql`):

**System Tables:**
- `global_config` - System-wide key-value settings (e.g., `ip_whitelist_enabled`)
- `app_config` - Application config (server, proxy settings)
- `ip_whitelist` - IP/CIDR whitelist entries
- `tls_config` - TLS certificates and auto-cert settings

**Routing Tables (v2):**
- `source` - Request sources (weixin, kdniao, etc.)
- `project` - Projects under each source
- `route_rule` - Routing rules with match_type, CEL expressions
- `plugin_registry` - Plugin registration (name, command path)

**Global Config Keys:**
- `ip_whitelist_enabled` - Controls IP whitelist feature on/off (values: "true"/"false")

### CEL Expression Engine

The CEL (Common Expression Language) engine provides flexible condition matching:

```go
// Available variables in expressions:
path    - map[string]string  // Path parameters: path.tenant == 'acme'
params  - map[string]string  // URL params: params.action == 'create'
headers - map[string]string  // Headers: headers['Content-Type'] == 'application/json'
body    - map[string]any     // JSON body: body['user']['role'] == 'admin'
method  - string             // HTTP method
host    - string             // Host header
pathStr - string             // Raw path string

// Example expressions:
"path.tenant == 'acme' && params.action == 'create'"
"headers['Authorization'].startsWith('Bearer')"
"body['user']['role'] == 'admin' || body['user']['role'] == 'superadmin'"
```

**Sandbox Security:**
- Expression size limit: 4096 bytes
- Execution timeout: 1000ms
- Forbidden patterns: import, exec, system, runtime, os.
- Allowed functions: equals, startsWith, endsWith, contains, has

### Plugin System

Plugins are independent processes communicating via gRPC:

```go
// RouterPlugin interface (plugin/interface.go)
type RouterPlugin interface {
    Info(ctx context.Context) (*PluginInfo, error)
    Match(ctx context.Context, req *MatchRequest) (*MatchResponse, error)
}
```

**Plugin Lifecycle:**
1. Prism scans plugin directories at startup
2. Each plugin runs as a separate process
3. Communication via gRPC protocol
4. Plugins are killed when Prism shuts down

See `docs/PLUGIN_DEVELOPMENT.md` for plugin development guide.

### Middleware Pattern

Middlewares are initialized via `middleware.Initialize()` called in `server.Run()`:

```go
middleware.Initialize(&middleware.Config{
    DefaultRPS:   100,
    DefaultBurst: 200,
})
```

### Global Configuration

System-wide settings are stored in `global_config` table and accessed via repository:

```go
// Get config value
config, err := queries.GetGlobalConfig(ctx, "ip_whitelist_enabled")
if config.Value == "true" {
    // Feature is enabled
}

// Set config value
err := queries.SetGlobalConfig(ctx, &db.SetGlobalConfigParams{
    Key:   "ip_whitelist_enabled",
    Value: "true",
})
```

**IP Whitelist Feature Toggle:**
- Key: `ip_whitelist_enabled`
- When disabled (value != "true"), the whitelist middleware passes all requests
- When enabled, only IPs in `ip_whitelist` table are allowed

### Application Configuration

Application config is stored in `app_config` table and accessed via `pkg/config`:

```go
import "github.com/rfancn/prism/pkg/config"

// Load all app config
configMgr := config.NewConfigManager()
appConfig, err := configMgr.LoadAppConfig(ctx)
// appConfig.Server.Host, appConfig.Server.Port, etc.

// Get/Set individual config
value, err := configMgr.GetConfig(ctx, "server.port")
err = configMgr.SetConfig(ctx, "server.port", "9090")
```

### Database Access

Use `repository.New()` to get a `*db.Queries` instance:

```go
queries := repository.New()
routes, err := queries.ListEnabledRoutes(ctx)
```

### Route Registration

Routes are loaded from database at startup in `router.NewRouter()`:

```go
// In router.LoadConfig()
config, err := r.loader.LoadAll(ctx)
// Routes are cached in r.config for fast matching
```

## Database Migrations

SQL schemas are in `assets/schema/`. After modifying:

1. Update schema files in `assets/schema/schema.sql`
2. Update query files in `assets/queries/`
3. Run `make generate` or `sqlc generate -f assets/sqlc.yaml` to regenerate code

## Important Notes

- Use `sdk.Logger()` for logging (not standard log package)
- Use `sdk.Db().My()` for database connection (wrapped by `repository.New()`)
- Config uses `mapstructure` tags
- Generated code lives in `autogen/` - never edit manually
- Pattern conversion: `{tenant}` in config -> `:tenant` for Gin routing
- Plugin binaries must be compiled with same Go version as main program
- Windows is not supported due to Go plugin limitations
- Default database path: `prism.db` (constant `g.DefaultDbPath`)