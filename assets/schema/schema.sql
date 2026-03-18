-- Schema for Prism HTTP/HTTPS request relay tool
-- SQLite database schema

-- ============================================================
-- 全局配置表：存储系统全局设置
-- ============================================================
CREATE TABLE IF NOT EXISTS global_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- IP whitelist table: stores allowed IP addresses/CIDR ranges
-- ============================================================
CREATE TABLE IF NOT EXISTS ip_whitelist (
    id TEXT PRIMARY KEY,
    ip_cidr TEXT NOT NULL UNIQUE,              -- IP address or CIDR range
    description TEXT,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- TLS configuration
-- ============================================================
CREATE TABLE IF NOT EXISTS tls_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER DEFAULT 0,
    cert_file TEXT,
    key_file TEXT,
    auto_cert INTEGER DEFAULT 0,
    domains TEXT,                              -- Comma-separated domains for auto-cert
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 1. 来源表：存储请求来源
-- ============================================================
CREATE TABLE IF NOT EXISTS source (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,          -- 来源名称，如 weixin, kdniao
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 2. 项目表：每个来源下可配置多个项目
-- ============================================================
CREATE TABLE IF NOT EXISTS project (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,            -- 关联来源
    name TEXT NOT NULL,                 -- 项目名称
    description TEXT,
    target_url TEXT,                    -- 目标URL
    priority INTEGER DEFAULT 0,         -- 项目优先级（数字越小优先级越高）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_id) REFERENCES source(id) ON DELETE CASCADE,
    UNIQUE(source_id, name)             -- 同一来源下项目名唯一
);

-- ============================================================
-- 3. 路由规则表
-- ============================================================
CREATE TABLE IF NOT EXISTS route_rule (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,           -- 关联项目
    name TEXT NOT NULL,                 -- 规则名称
    match_type TEXT NOT NULL CHECK (match_type IN (
        'param_path',      -- 参数化路径匹配
        'url_param',       -- URL参数匹配
        'request_body',    -- 请求内容匹配
        'request_form',    -- 请求表单匹配
        'plugin'           -- 插件模式
    )),

    -- 参数化路径匹配专用字段
    path_pattern TEXT,                  -- 带参数的路径，如 /users/{id}/orders/{orderId}

    -- CEL表达式（用于所有内置模式）
    cel_expression TEXT,                -- CEL表达式

    -- 插件模式专用字段
    plugin_name TEXT,                   -- 插件名称

    priority INTEGER DEFAULT 0,         -- 匹配优先级（数字越小优先级越高）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    UNIQUE(project_id, name)            -- 同一项目下规则名唯一
);

-- ============================================================
-- 4. 插件注册表
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_registry (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,          -- 插件名称
    description TEXT,                   -- 插件描述
    version TEXT,                       -- 插件版本
    command TEXT NOT NULL,              -- 插件可执行文件路径或命令
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 索引
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_whitelist_ip ON ip_whitelist(ip_cidr);
CREATE INDEX IF NOT EXISTS idx_whitelist_enabled ON ip_whitelist(enabled);
CREATE INDEX IF NOT EXISTS idx_project_source_id ON project(source_id);
CREATE INDEX IF NOT EXISTS idx_project_priority ON project(priority);
CREATE INDEX IF NOT EXISTS idx_route_rule_project_id ON route_rule(project_id);
CREATE INDEX IF NOT EXISTS idx_route_rule_priority ON route_rule(priority);
CREATE INDEX IF NOT EXISTS idx_plugin_registry_name ON plugin_registry(name);

-- ============================================================
-- 应用配置表：存储应用级别的配置项
-- ============================================================
CREATE TABLE IF NOT EXISTS app_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL DEFAULT 'string',
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);