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

# Run tests (all packages)
go test ./...

# Run tests with verbose output
go test -v ./...

# Generate sqlc code (after modifying SQL schema/queries)
go generate ./...

# Run linter
go vet ./...

# Run the service
./prism run -c prism.toml

# Run TUI management interface
./prism tui
```

**Note**: 在国内网络环境下，需要设置 Go 代理：
```bash
export GOPROXY=https://goproxy.cn,direct
```

## Key Dependencies

- **SDK**: `github.com/hdget/sdk v0.5.1` - provides Logger and DB access via `sdk.Logger()` and `sdk.Db()`
- **HTTP Framework**: `github.com/gin-gonic/gin` - HTTP server and routing
- **Database**: SQLite with sqlc code generation (`github.com/hdget/sdk/providers/db/sqlite3/sqlc v0.0.2`)
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

- `cmd/` - Cobra CLI commands (`run`, `tui`, `route`, `apikey`, `version`, `migrations`)
- `g/` - Global configuration structures (`g.Config`) and constants
- `pkg/` - Core packages:
  - `router/` - Three-layer routing, matching, and forwarding
  - `matcher/` - Request matcher implementations (5 types)
  - `cel/` - CEL expression engine with sandbox security
  - `middleware/` - Gin middleware (registration pattern with `Register()`/`Get()`)
  - `proxy/` - Reverse proxy with director-based request rewriting
  - `server/` - Gin HTTP server setup and route registration
  - `monitor/` - Prometheus metrics and health endpoints
  - `types/` - Shared types across packages (TLS config)
- `tui/` - Bubble Tea TUI application (models for routes, apikeys, whitelist)
- `repository/` - Data access layer using `repository.New()` for `db.Queries`
- `plugin/` - Plugin system implementation
  - `interface.go` - RouterPlugin interface definition
  - `manager.go` - Plugin lifecycle management
  - `client.go` / `server.go` - gRPC communication
- `assets/` - SQL schema (`schema/`) and queries (`queries/`) for sqlc
- `autogen/db/` - sqlc-generated code (do not edit manually)
- `docs/` - Documentation files

### Configuration Structure

Config is split into `[sdk]` and `[app]` sections in TOML:

- `[sdk]` - SDK-level config (logging, database)
- `[app]` - Application config (server, proxy, routes, rate limiting)

Access via `g.Config` global variable after loading. Config structs use `mapstructure` tags for TOML binding.

### Database Schema

Key tables (see `assets/schema/schema.sql`):

**Legacy Tables:**
- `route` - Legacy routing patterns (deprecated, use route_rule)
- `ip_whitelist` - IP/CIDR whitelist entries
- `api_key` - API keys with user_id for rate limiting
- `tls_config` - TLS certificates and auto-cert settings

**New Tables (v2):**
- `source` - Request sources (weixin, kdniao, etc.)
- `project` - Projects under each source
- `route_rule` - Routing rules with match_type, CEL expressions
- `plugin_registry` - Plugin registration (name, command path)
- `header` - Custom headers per route rule (cascade delete)

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
- Config uses `mapstructure` tags for TOML binding
- Generated code lives in `autogen/` - never edit manually
- Pattern conversion: `{tenant}` in config -> `:tenant` for Gin routing
- Plugin binaries must be compiled with same Go version as main program
- Windows is not supported due to Go plugin limitations