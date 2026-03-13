# Prism — HTTP/HTTPS 请求中继构建计划

## 项目概述

**目标：** 构建一个 HTTP/HTTPS 请求中继工具，支持透明转发、基于标识符的路由、TUI 管理、Header 注入、IP 白名单、API Key 认证、按用户限流、Prometheus 监控和增强日志。

**技术栈：**
- 语言：Go 1.21+
- SDK：`github.com/hdget/sdk v0.5.1`
- HTTP 框架：Gin (`github.com/gin-gonic/gin`)
- 数据库：SQLite + sqlc 代码生成
- CLI：Cobra (`github.com/spf13/cobra`)
- 配置：TOML (`github.com/hdget/sdk/providers/config/konaf`)
- TUI：Bubble Tea (`github.com/charmbracelet/bubbletea`)
- 监控：Prometheus `client_golang`
- 限流：`golang.org/x/time/rate`

**模式：** 直接模式（无 GitHub CLI，使用原地编辑工作流）

---

## 目录结构

参考 `h1dian/cmc/backend` 项目结构：

```
prism/
├── assets/                    # sqlc 相关资源
│   ├── schema/               # SQL Schema 定义
│   │   ├── route.sql
│   │   ├── header.sql
│   │   ├── whitelist.sql
│   │   ├── api_key.sql
│   │   └── rate_limit.sql
│   ├── queries/              # SQL 查询定义
│   │   ├── route.sql
│   │   ├── header.sql
│   │   ├── whitelist.sql
│   │   ├── api_key.sql
│   │   └── rate_limit.sql
│   ├── migrations/           # 数据库迁移文件
│   ├── sqlc.yaml             # sqlc 配置
│   └── assets.go
├── autogen/                   # 自动生成的代码
│   └── db/
│       ├── db.go
│       ├── models.go
│       ├── querier.go
│       ├── route.sql.go
│       ├── header.sql.go
│       ├── whitelist.sql.go
│       ├── api_key.sql.go
│       └── rate_limit.sql.go
├── cmd/                       # CLI 命令
│   ├── root.go
│   ├── run.go                 # 启动代理服务
│   ├── tui.go                 # 启动 TUI
│   ├── route.go               # 路由管理命令
│   ├── apikey.go              # API Key 管理命令
│   └── version.go
├── controller/                # 控制器（管理 API）
│   └── v1/
│       ├── route.go
│       ├── apikey.go
│       └── stats.go
├── g/                         # 全局配置和变量
│   ├── global.go
│   └── config.go
├── pkg/                       # 公共包
│   ├── middleware/
│   │   ├── middleware.go
│   │   ├── mdw_whitelist.go
│   │   ├── mdw_auth.go
│   │   ├── mdw_ratelimit.go
│   │   ├── mdw_logger.go
│   │   └── mdw_proxy.go
│   ├── parser/               # 标识符解析器
│   │   ├── parser.go
│   │   ├── path.go
│   │   ├── json_body.go
│   │   └── url_param.go
│   ├── proxy/                # 反向代理核心
│   │   ├── proxy.go
│   │   ├── director.go
│   │   └── target_tls.go
│   ├── ratelimit/            # 限流器
│   │   └── limiter.go
│   ├── monitor/              # Prometheus 指标
│   │   ├── metrics.go
│   │   └── handler.go
│   └── server/               # HTTP 服务器
│       ├── server.go
│       └── server_web.go
├── repository/                # 数据访问层
│   ├── route.go
│   ├── header.go
│   ├── whitelist.go
│   ├── api_key.go
│   └── rate_limit.go
├── service/                   # 业务逻辑层
│   ├── route.go
│   ├── header.go
│   ├── whitelist.go
│   ├── api_key.go
│   └── ratelimit.go
├── tui/                       # TUI 界面
│   ├── app.go
│   ├── styles.go
│   ├── components/
│   ├── routes/
│   ├── headers/
│   ├── whitelist/
│   ├── apikeys/
│   ├── ratelimit/
│   └── tls/
├── main.go                    # 入口文件
├── prism.example.toml         # 示例配置文件
├── go.mod
└── go.sum
```

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                         Prism 架构图                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐    ┌──────────────────────────────────────────┐   │
│  │   客户端  │───▶│  Gin HTTP Server                         │   │
│  │ (HTTP/S) │    │  :8080 (HTTP) / :8443 (HTTPS)            │   │
│  └──────────┘    └─────────────────┬────────────────────────┘   │
│                                    │                             │
│                                    ▼                             │
│                    ┌───────────────────────────────┐             │
│                    │  Gin 中间件链                   │             │
│                    │  1. Logger (sdk.Logger)       │             │
│                    │  2. IP Whitelist              │             │
│                    │  3. API Key Auth              │             │
│                    │  4. Rate Limit (per user)     │             │
│                    │  5. Proxy Handler             │             │
│                    └───────────────┬───────────────┘             │
│                                    │                             │
│                                    ▼                             │
│                    ┌───────────────────────────────┐             │
│                    │  标识符解析器                   │             │
│                    │  - Path / JSON Body / URL     │             │
│                    └───────────────┬───────────────┘             │
│                                    │                             │
│                                    ▼                             │
│                    ┌───────────────────────────────┐             │
│                    │  路由匹配 & Header 注入         │             │
│                    │  (SQLite + sqlc)              │             │
│                    └───────────────┬───────────────┘             │
│                                    │                             │
│                                    ▼                             │
│  ┌──────────┐    ┌──────────────────────────────────────────┐   │
│  │   目标    │◀───│  httputil.ReverseProxy                   │   │
│  │ (HTTP/S) │    └──────────────────────────────────────────┘   │
│  └──────────┘                                                    │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    监控端点                                 │   │
│  │  GET /metrics   - Prometheus 指标                         │   │
│  │  GET /health    - 健康检查                                │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    CLI (Cobra)                             │   │
│  │  prism run       启动代理服务                             │   │
│  │  prism tui       启动 TUI 管理界面                        │   │
│  │  prism route     路由管理                                 │   │
│  │  prism apikey    API Key 管理                             │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    TUI 管理界面 (Bubble Tea)               │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 步骤 1：项目初始化 & 目录结构

**依赖：** 无（基础步骤）

**模型层级：** 默认

**上下文简介：**
初始化 Go 模块，创建项目目录结构，添加核心依赖。

**任务清单：**
- [ ] 初始化 Go 模块：`go mod init github.com/rfancn/prism`
- [ ] 添加核心依赖：
  ```bash
  go get github.com/hdget/sdk@v0.5.1
  go get github.com/gin-gonic/gin
  go get github.com/spf13/cobra
  go get github.com/charmbracelet/bubbletea
  go get github.com/charmbracelet/lipgloss
  go get github.com/charmbracelet/bubbles
  go get github.com/prometheus/client_golang
  go get golang.org/x/time/rate
  ```
- [ ] 创建目录结构
- [ ] 创建 `g/global.go`：
  ```go
  package g

  const (
      App = "prism"
  )

  var (
      Debug bool
  )
  ```
- [ ] 创建 `g/config.go`：配置结构定义
- [ ] 创建 `main.go`：
  ```go
  package main

  import "github.com/rfancn/prism/cmd"

  //go:generate sqlc generate -f assets/sqlc.yaml
  func main() {
      cmd.Execute()
  }
  ```

**创建文件：**
- `go.mod`
- `g/global.go`
- `g/config.go`
- `main.go`
- `assets/assets.go`

**验证命令：**
```bash
go mod tidy
go build ./...
```

**退出标准：**
- 项目编译无错误
- 目录结构创建完成

**回滚方案：** 删除创建的文件和目录

---

## 步骤 2：SQL Schema & sqlc 配置

**依赖：** 步骤 1

**模型层级：** 默认

**上下文简介：**
创建数据库 Schema 和 sqlc 配置，生成类型安全的数据库访问代码。

**任务清单：**
- [ ] 创建 `assets/sqlc.yaml`：
  ```yaml
  version: "2"
  sql:
    - schema: "schema"
      queries: "queries"
      engine: "sqlite"
      gen:
        go:
          package: "db"
          sql_package: "database/sql"
          out: "../autogen/db"
          emit_json_tags: true
          emit_empty_slices: true
          emit_interface: true
          emit_result_struct_pointers: true
          emit_params_struct_pointers: true
  ```
- [ ] 创建 `assets/schema/route.sql`：路由表 Schema
- [ ] 创建 `assets/schema/header.sql`：Header 表 Schema
- [ ] 创建 `assets/schema/whitelist.sql`：IP 白名单表 Schema
- [ ] 创建 `assets/schema/api_key.sql`：API Key 表 Schema
- [ ] 创建 `assets/schema/rate_limit.sql`：限流配置表 Schema
- [ ] 创建 `assets/schema/tls.sql`：TLS 配置表 Schema
- [ ] 创建 `assets/queries/` 下的查询文件
- [ ] 安装 sqlc：`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [ ] 生成代码：`go generate ./...`

**创建文件：**
- `assets/sqlc.yaml`
- `assets/schema/*.sql`
- `assets/queries/*.sql`
- `autogen/db/*.go`（生成）

**验证命令：**
```bash
go generate ./...
go build ./...
```

**退出标准：**
- sqlc 代码生成成功
- 生成的代码编译通过

**回滚方案：** 删除 `assets/schema/`, `assets/queries/`, `autogen/` 目录

---

## 步骤 3：配置管理 (TOML)

**依赖：** 步骤 1

**模型层级：** 默认

**上下文简介：**
创建 TOML 配置结构，参考 gateway.test.toml 格式。配置分为 `[sdk]` 和 `[app]` 两部分。

**任务清单：**
- [ ] 更新 `g/config.go`：
  ```go
  package g

  type RootConfig struct {
      Sdk SdkConfig `mapstructure:"sdk"`
      App AppConfig `mapstructure:"app"`
  }

  type SdkConfig struct {
      Log    LogConfig    `mapstructure:"log"`
      Sqlite SqliteConfig `mapstructure:"sqlite"`
  }

  type LogConfig struct {
      Level    string       `mapstructure:"level"`
      Filename string       `mapstructure:"filename"`
      Rotate   RotateConfig `mapstructure:"rotate"`
  }

  type RotateConfig struct {
      MaxAge       int `mapstructure:"max_age"`
      RotationTime int `mapstructure:"rotation_time"`
  }

  type SqliteConfig struct {
      Db string `mapstructure:"db"`
  }

  type AppConfig struct {
      Server    ServerConfig    `mapstructure:"server"`
      Proxy     ProxyConfig     `mapstructure:"proxy"`
      Route     RouteConfig     `mapstructure:"route"`
      RateLimit RateLimitConfig `mapstructure:"ratelimit"`
  }

  type ServerConfig struct {
      Host    string `mapstructure:"host"`
      Port    int    `mapstructure:"port"`
      TLSPort int    `mapstructure:"tls_port"`
  }

  type ProxyConfig struct {
      ReadTimeout  int `mapstructure:"read_timeout"`
      WriteTimeout int `mapstructure:"write_timeout"`
      IdleTimeout  int `mapstructure:"idle_timeout"`
  }

  type RouteConfig struct {
      ProtectPrefix string `mapstructure:"protect_prefix"`
      PublicPrefix  string `mapstructure:"public_prefix"`
  }

  type RateLimitConfig struct {
      WindowSize int `mapstructure:"window_size"`
      Limit      int `mapstructure:"limit"`
  }
  ```
- [ ] 创建 `prism.example.toml`：
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

      [app.route]
          protect_prefix = "/api"
          public_prefix = "/public"

      [app.ratelimit]
          window_size = 10
          limit = 100
  ```

**创建文件：**
- `g/config.go`
- `prism.example.toml`

**验证命令：**
```bash
go build ./...
```

**退出标准：**
- 配置结构定义完整
- 示例配置文件创建

**回滚方案：** 删除 `g/config.go`, `prism.example.toml`

---

## 步骤 4：CLI 命令框架 (Cobra)

**依赖：** 步骤 1, 步骤 3

**模型层级：** 默认

**上下文简介：**
创建 Cobra CLI 命令框架，支持 run、tui、route、apikey、version 等子命令。

**任务清单：**
- [ ] 创建 `cmd/root.go`：
  ```go
  package cmd

  import (
      "github.com/rfancn/prism/g"
      panicUtils "github.com/hdget/utils/panic"
      "github.com/hdget/utils/logger"
      "github.com/spf13/cobra"
  )

  var rootCmd = &cobra.Command{
      Long:  "Prism - HTTP/HTTPS 请求中继工具",
      Short: "HTTP/HTTPS 请求中继工具",
      Use:   "prism",
  }

  func init() {
      rootCmd.PersistentFlags().BoolVarP(&g.Debug, "debug", "", true, "debug mode")
      rootCmd.AddCommand(cmdRun)
      rootCmd.AddCommand(cmdTui)
      rootCmd.AddCommand(cmdRoute)
      rootCmd.AddCommand(cmdApikey)
      rootCmd.AddCommand(cmdVersion)
  }

  func Execute() {
      defer func() {
          if r := recover(); r != nil {
              panicUtils.RecordErrorStack(g.App)
          }
      }()

      if err := rootCmd.Execute(); err != nil {
          logger.Fatal("root command execute", "err", err)
      }
  }
  ```
- [ ] 创建 `cmd/run.go`：启动代理服务
- [ ] 创建 `cmd/tui.go`：启动 TUI 界面
- [ ] 创建 `cmd/route.go`：路由管理命令
- [ ] 创建 `cmd/apikey.go`：API Key 管理命令
- [ ] 创建 `cmd/version.go`：版本信息命令

**创建文件：**
- `cmd/root.go`
- `cmd/run.go`
- `cmd/tui.go`
- `cmd/route.go`
- `cmd/apikey.go`
- `cmd/version.go`

**验证命令：**
```bash
go build -o prism .
./prism --help
./prism version
```

**退出标准：**
- CLI 命令正常工作
- 子命令正确注册

**回滚方案：** 删除 `cmd/` 目录

---

## 步骤 5：Repository 层

**依赖：** 步骤 2

**模型层级：** 默认

**上下文简介：**
创建数据访问层，封装 sqlc 生成的数据库操作。

**任务清单：**
- [ ] 创建 `repository/route.go`：路由数据访问
- [ ] 创建 `repository/header.go`：Header 数据访问
- [ ] 创建 `repository/whitelist.go`：白名单数据访问
- [ ] 创建 `repository/api_key.go`：API Key 数据访问
- [ ] 创建 `repository/rate_limit.go`：限流配置数据访问

**创建文件：**
- `repository/route.go`
- `repository/header.go`
- `repository/whitelist.go`
- `repository/api_key.go`
- `repository/rate_limit.go`

**验证命令：**
```bash
go build ./...
go test ./repository/... -v
```

**退出标准：**
- Repository 层编译通过
- 单元测试通过

**回滚方案：** 删除 `repository/` 目录

---

## 步骤 6：标识符解析器

**依赖：** 步骤 1

**模型层级：** 默认

**上下文简介：**
实现从三种来源提取请求标识符：路径参数、JSON Body 字段、URL 查询参数。

**任务清单：**
- [ ] 创建 `pkg/parser/parser.go`：Parser 接口
- [ ] 创建 `pkg/parser/path.go`：路径参数解析
- [ ] 创建 `pkg/parser/json_body.go`：JSON Body 解析
- [ ] 创建 `pkg/parser/url_param.go`：URL 参数解析
- [ ] 创建 `pkg/parser/parser_test.go`

**创建文件：**
- `pkg/parser/parser.go`
- `pkg/parser/path.go`
- `pkg/parser/json_body.go`
- `pkg/parser/url_param.go`
- `pkg/parser/parser_test.go`

**验证命令：**
```bash
go test ./pkg/parser/... -v -cover
```

**退出标准：**
- 三种解析器均已实现
- 测试覆盖率 > 80%

**回滚方案：** 删除 `pkg/parser/` 目录

---

## 步骤 7：Gin 中间件

**依赖：** 步骤 2, 步骤 5

**模型层级：** 默认

**上下文简介：**
创建 Gin 中间件：IP 白名单、API Key 认证、按用户限流、日志记录。

**任务清单：**
- [ ] 创建 `pkg/middleware/middleware.go`：中间件注册和获取
- [ ] 创建 `pkg/middleware/mdw_logger.go`：日志中间件（使用 sdk.Logger()）
- [ ] 创建 `pkg/middleware/mdw_whitelist.go`：IP 白名单中间件
- [ ] 创建 `pkg/middleware/mdw_auth.go`：API Key 认证中间件
- [ ] 创建 `pkg/middleware/mdw_ratelimit.go`：按用户限流中间件
- [ ] 创建 `pkg/middleware/mdw_proxy.go`：代理处理中间件

**创建文件：**
- `pkg/middleware/middleware.go`
- `pkg/middleware/mdw_logger.go`
- `pkg/middleware/mdw_whitelist.go`
- `pkg/middleware/mdw_auth.go`
- `pkg/middleware/mdw_ratelimit.go`
- `pkg/middleware/mdw_proxy.go`

**验证命令：**
```bash
go build ./...
go test ./pkg/middleware/... -v
```

**退出标准：**
- 中间件编译通过
- 单元测试通过

**回滚方案：** 删除 `pkg/middleware/` 目录

---

## 步骤 8：反向代理核心

**依赖：** 步骤 1, 步骤 6

**模型层级：** 默认

**上下文简介：**
实现核心反向代理，基于 `net/http/httputil.ReverseProxy`，集成标识符解析和 Header 注入。

**任务清单：**
- [ ] 创建 `pkg/proxy/proxy.go`：代理处理器
- [ ] 创建 `pkg/proxy/director.go`：请求重写和 Header 注入
- [ ] 创建 `pkg/proxy/target_tls.go`：目标 TLS 配置

**创建文件：**
- `pkg/proxy/proxy.go`
- `pkg/proxy/director.go`
- `pkg/proxy/target_tls.go`

**验证命令：**
```bash
go build ./...
go test ./pkg/proxy/... -v
```

**退出标准：**
- 代理转发正常
- 单元测试通过

**回滚方案：** 删除 `pkg/proxy/` 目录

---

## 步骤 9：限流器

**依赖：** 步骤 1

**模型层级：** 默认

**上下文简介：**
实现基于令牌桶算法的按用户限流器。

**任务清单：**
- [ ] 创建 `pkg/ratelimit/limiter.go`：用户限流器管理

**创建文件：**
- `pkg/ratelimit/limiter.go`

**验证命令：**
```bash
go build ./...
go test ./pkg/ratelimit/... -v
```

**退出标准：**
- 限流器正常工作

**回滚方案：** 删除 `pkg/ratelimit/` 目录

---

## 步骤 10：Prometheus 监控

**依赖：** 步骤 1

**模型层级：** 默认

**上下文简介：**
实现 Prometheus 监控指标和健康检查端点。

**任务清单：**
- [ ] 创建 `pkg/monitor/metrics.go`：指标定义
- [ ] 创建 `pkg/monitor/handler.go`：/metrics 和 /health 端点

**创建文件：**
- `pkg/monitor/metrics.go`
- `pkg/monitor/handler.go`

**验证命令：**
```bash
go build ./...
```

**退出标准：**
- 监控端点正常

**回滚方案：** 删除 `pkg/monitor/` 目录

---

## 步骤 11：HTTP 服务器 (Gin)

**依赖：** 步骤 3, 步骤 7, 步骤 8, 步骤 10

**模型层级：** 默认

**上下文简介：**
创建 Gin HTTP 服务器，组装中间件链，注册路由。

**任务清单：**
- [ ] 创建 `pkg/server/server.go`：Server 接口
- [ ] 创建 `pkg/server/server_web.go`：
  ```go
  package server

  import (
      "github.com/rfancn/prism/g"
      "github.com/rfancn/prism/pkg/middleware"
      "context"
      "fmt"
      "os"
      "os/signal"
      "sync"
      "syscall"
      "time"

      "github.com/gin-gonic/gin"
      "github.com/hdget/sdk"
      panicUtils "github.com/hdget/utils/panic"
  )

  type webServerImpl struct {
      engine *gin.Engine
      ctx    context.Context
      cancel context.CancelFunc
      wg     *sync.WaitGroup
  }

  const gracefulShutdownTime = 15 * time.Second

  var commonMiddlewares = []string{
      middleware.NameLogger,
  }

  func New(address string) (Server, error) {
      engine := gin.New()

      // 初始化中间件链
      // ...

      wg := &sync.WaitGroup{}
      ctx, cancel := context.WithCancel(context.Background())
      return &webServerImpl{
          engine: engine,
          ctx:    ctx,
          cancel: cancel,
          wg:     wg,
      }, nil
  }

  func (w *webServerImpl) Run() {
      // 监听信号并启动服务器
  }
  ```

**创建文件：**
- `pkg/server/server.go`
- `pkg/server/server_web.go`

**验证命令：**
```bash
go build ./...
```

**退出标准：**
- 服务器正常启动
- 中间件链正确

**回滚方案：** 删除 `pkg/server/` 目录

---

## 步骤 12：Service 层

**依赖：** 步骤 5

**模型层级：** 默认

**上下文简介：**
创建业务逻辑层，封装复杂业务操作。

**任务清单：**
- [ ] 创建 `service/route.go`
- [ ] 创建 `service/header.go`
- [ ] 创建 `service/whitelist.go`
- [ ] 创建 `service/api_key.go`
- [ ] 创建 `service/ratelimit.go`

**创建文件：**
- `service/route.go`
- `service/header.go`
- `service/whitelist.go`
- `service/api_key.go`
- `service/ratelimit.go`

**验证命令：**
```bash
go build ./...
```

**退出标准：**
- Service 层编译通过

**回滚方案：** 删除 `service/` 目录

---

## 步骤 13：Controller 层

**依赖：** 步骤 12

**模型层级：** 默认

**上下文简介：**
创建管理 API 控制器，供 TUI 或外部工具调用。

**任务清单：**
- [ ] 创建 `controller/v1/route.go`：路由管理 API
- [ ] 创建 `controller/v1/apikey.go`：API Key 管理 API
- [ ] 创建 `controller/v1/stats.go`：统计信息 API

**创建文件：**
- `controller/v1/route.go`
- `controller/v1/apikey.go`
- `controller/v1/stats.go`

**验证命令：**
```bash
go build ./...
```

**退出标准：**
- Controller 层编译通过

**回滚方案：** 删除 `controller/` 目录

---

## 步骤 14：TUI 界面 (Bubble Tea)

**依赖：** 步骤 1, 步骤 12

**模型层级：** 默认

**上下文简介：**
使用 Bubble Tea 创建 TUI 管理界面。

**任务清单：**
- [ ] 创建 `tui/app.go`：主应用
- [ ] 创建 `tui/styles.go`：样式定义
- [ ] 创建 `tui/components/`：通用组件
- [ ] 创建 `tui/routes/`：路由管理界面
- [ ] 创建 `tui/headers/`：Header 管理界面
- [ ] 创建 `tui/whitelist/`：白名单管理界面
- [ ] 创建 `tui/apikeys/`：API Key 管理界面
- [ ] 创建 `tui/ratelimit/`：限流配置界面
- [ ] 创建 `tui/tls/`：TLS 配置界面

**创建文件：**
- `tui/app.go`
- `tui/styles.go`
- `tui/components/*.go`
- `tui/routes/*.go`
- `tui/headers/*.go`
- `tui/whitelist/*.go`
- `tui/apikeys/*.go`
- `tui/ratelimit/*.go`
- `tui/tls/*.go`

**验证命令：**
```bash
go build ./...
```

**退出标准：**
- TUI 界面可用

**回滚方案：** 删除 `tui/` 目录

---

## 步骤 15：集成 & 测试

**依赖：** 所有前置步骤

**模型层级：** 默认

**上下文简介：**
集成所有组件，创建完整测试，添加文档。

**任务清单：**
- [ ] 更新 `cmd/run.go`：完整的初始化流程
- [ ] 创建集成测试
- [ ] 创建 `README.md`
- [ ] 创建 `docs/` 目录
- [ ] 添加 Makefile

**创建文件：**
- `tests/integration/integration_test.go`
- `README.md`
- `docs/`
- `Makefile`

**验证命令：**
```bash
make test
make build
./prism run -c prism.example.toml
```

**退出标准：**
- 所有测试通过
- 文档完整
- 服务正常运行

**回滚方案：** 根据需要删除文件

---

## 依赖关系图

```
步骤 1 (初始化)
   │
   ├── 步骤 2 (sqlc) ────┬── 步骤 5 (Repository) ──┬── 步骤 12 (Service) ── 步骤 13 (Controller)
   │                      │                         │
   │                      │                         └── 步骤 14 (TUI)
   │                      │
   ├── 步骤 3 (配置) ─────┼── 步骤 4 (CLI)
   │                      │
   ├── 步骤 6 (解析器) ───┼── 步骤 7 (中间件) ──┐
   │                      │                     │
   ├── 步骤 8 (代理) ─────┤                     ├── 步骤 11 (服务器)
   │                      │                     │
   ├── 步骤 9 (限流器) ───┤                     │
   │                      │                     │
   └── 步骤 10 (监控) ────┴─────────────────────┘
                                                          │
                                         步骤 15 (集成) ───┘
```

**可并行执行的步骤：**
- 步骤 1 完成后：步骤 2, 3, 6, 8, 9, 10 可并行
- 步骤 2 完成后：步骤 5 开始
- 步骤 5 完成后：步骤 12 开始
- 步骤 7, 8, 9, 10 完成后：步骤 11 开始

---

## 不变量（每个步骤后验证）

1. **构建：** `go build ./...` 成功
2. **测试：** `go test ./...` 通过
3. **检查：** `go vet ./...` 通过
4. **无回归：** 之前工作的功能仍然正常

---

## 回滚协议

每个步骤包含具体的回滚说明。通用方法：
1. 确定要回滚到的步骤
2. 删除后续步骤创建的文件
3. 运行 `go mod tidy`
4. 验证构建成功
5. 运行测试

---

## 成功标准

- [ ] 代理透明转发 HTTP/HTTPS 请求
- [ ] 标识符解析支持所有三种类型
- [ ] 可通过 CLI/TUI 管理路由
- [ ] Header 正确注入
- [ ] IP 白名单阻止非允许的 IP
- [ ] API Key 认证正常工作
- [ ] 按用户限流正常工作
- [ ] Prometheus 指标正常暴露
- [ ] 健康检查端点正常
- [ ] sdk.Logger() 日志正常输出
- [ ] sdk.Db() 数据库访问正常
- [ ] TOML 配置加载正常
- [ ] Cobra CLI 命令正常工作
- [ ] Gin HTTP 服务器正常工作
- [ ] TUI 界面可用
- [ ] 所有测试通过
- [ ] 文档完整

---

## 技术依赖汇总

```go
// go.mod 核心依赖
require (
    github.com/hdget/sdk v0.5.1
    github.com/gin-gonic/gin latest
    github.com/spf13/cobra latest
    github.com/charmbracelet/bubbletea latest
    github.com/charmbracelet/lipgloss latest
    github.com/charmbracelet/bubbles latest
    github.com/prometheus/client_golang latest
    golang.org/x/time latest
)
```