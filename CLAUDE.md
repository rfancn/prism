# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Prism is an HTTP/HTTPS request relay tool written in Go. It supports transparent forwarding, identifier-based routing, TUI management, header injection, IP whitelisting, API key authentication, and per-user rate limiting.

## Build & Development Commands

```bash
# Build the binary
go build -o prism .
# or
make build

# Run tests
go test ./...
# or
make test

# Generate sqlc code (after modifying SQL schema/queries)
go generate ./...
# or
make generate

# Run linter
go vet ./...
# or
make lint

# Run the service
./prism run -c prism.toml

# Run TUI management interface
./prism tui
```

## Key Dependencies

- **SDK**: `github.com/hdget/sdk v0.5.1` - provides Logger and DB access
- **HTTP Framework**: `github.com/gin-gonic/gin`
- **Database**: SQLite with sqlc code generation (`github.com/hdget/sdk/providers/db/sqlite3/sqlc v0.0.2`)
- **CLI**: `github.com/spf13/cobra`
- **Monitoring**: `github.com/prometheus/client_golang`
- **Rate Limiting**: `golang.org/x/time/rate`

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Request Flow                             │
├─────────────────────────────────────────────────────────────────┤
│  Client ──▶ Gin Server ──▶ Middleware Chain ──▶ Proxy Handler   │
│                                │                                 │
│            Routes loaded from DB at startup (server.New())       │
│            Pattern {tenant} → :tenant for Gin                    │
│                                │                                 │
│                     ┌──────────┴──────────┐                      │
│                     │   Proxy Handler     │                      │
│                     │   with middleware:  │                      │
│                     │   1. Logger         │                      │
│                     │   2. IP Whitelist   │                      │
│                     │   3. API Key Auth   │                      │
│                     │   4. Rate Limit     │                      │
│                     └──────────┬──────────┘                      │
│                                │                                 │
│            ┌───────────────────┼───────────────────┐             │
│            │                   │                   │             │
│      c.Param()          c.Query()          JSON Body            │
│      (path params)      (URL params)       (nested fields)      │
│            │                   │                   │             │
│            └───────────────────┼───────────────────┘             │
│                                │                                 │
│                     Header Inject + Reverse Proxy                │
│                                │                                 │
│                                ▼                                 │
│                     httputil.ReverseProxy ──▶ Target             │
└─────────────────────────────────────────────────────────────────┘
```

### Directory Structure

- `cmd/` - Cobra CLI commands (`run`, `tui`, `route`, `apikey`, `version`)
- `g/` - Global configuration structures and constants
- `pkg/` - Core packages:
  - `middleware/` - Gin middleware (registration pattern with `Register()`/`Get()`)
  - `parser/` - JSON body identifier extraction (nested field support)
  - `proxy/` - Reverse proxy with director-based request rewriting
  - `server/` - Gin HTTP server setup and route registration
  - `monitor/` - Prometheus metrics and health endpoints
  - `types/` - Shared types across packages
- `repository/` - Data access layer using `repository.New()` for `db.Queries`
- `assets/` - SQL schema and queries for sqlc
- `autogen/db/` - sqlc-generated code (do not edit manually)

### Configuration Structure

Config is split into `[sdk]` and `[app]` sections in TOML:

- `[sdk]` - SDK-level config (logging, database)
- `[app]` - Application config (server, proxy, routes, rate limiting)

Access via `g.Config` global variable after loading.

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

### Identifier Extraction

Three extraction sources supported:
- **Path**: Uses Gin's built-in `c.Param()` - pattern `/api/{tenant}/users` is converted to `/api/:tenant/users`
- **URL Param**: Uses Gin's built-in `c.Query()` for query parameter extraction
- **JSON Body**: Custom parser in `pkg/parser/` supports dot notation for nested fields (e.g., `user.tenant_id`)

### Route Registration

Routes are loaded from database at startup in `server.New()` and registered directly with Gin:

```go
// In server.setupRoutes()
routes, _ := queries.ListEnabledRoutes(ctx)
for _, route := range routes {
    ginPattern := convertPattern(route.Pattern) // {tenant} → :tenant
    s.engine.Any(ginPattern, handler)
}
```

## Database Migrations

SQL schemas are in `assets/schema/`. After modifying:

1. Update schema files in `assets/schema/`
2. Update query files in `assets/queries/`
3. Run `make generate` to regenerate sqlc code

## Important Notes

- Use `sdk.Logger()` for logging (not standard log package)
- Use `sdk.Db()` for database access
- Config uses `mapstructure` tags for TOML binding
- Generated code lives in `autogen/` - never edit manually