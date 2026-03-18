-- 初始化数据库schema

-- ============================================================
-- IP白名单表
-- ============================================================
CREATE TABLE IF NOT EXISTS ip_whitelist (
    id TEXT PRIMARY KEY,
    ip_cidr TEXT NOT NULL UNIQUE,
    description TEXT,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- TLS配置表
-- ============================================================
CREATE TABLE IF NOT EXISTS tls_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER DEFAULT 0,
    cert_file TEXT,
    key_file TEXT,
    auto_cert INTEGER DEFAULT 0,
    domains TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 来源表
-- ============================================================
CREATE TABLE IF NOT EXISTS source (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 项目表
-- ============================================================
CREATE TABLE IF NOT EXISTS project (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_id) REFERENCES source(id) ON DELETE CASCADE,
    UNIQUE(source_id, name)
);

-- ============================================================
-- 路由规则表
-- ============================================================
CREATE TABLE IF NOT EXISTS route_rule (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    match_type TEXT NOT NULL CHECK (match_type IN (
        'param_path',
        'url_param',
        'request_body',
        'request_form',
        'plugin'
    )),
    path_pattern TEXT,
    cel_expression TEXT,
    plugin_name TEXT,
    target_url TEXT NOT NULL,
    enabled INTEGER DEFAULT 1,
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    UNIQUE(project_id, name)
);

-- ============================================================
-- 插件注册表
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_registry (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    version TEXT,
    command TEXT NOT NULL,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 索引
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_whitelist_ip ON ip_whitelist(ip_cidr);
CREATE INDEX IF NOT EXISTS idx_whitelist_enabled ON ip_whitelist(enabled);
CREATE INDEX IF NOT EXISTS idx_source_enabled ON source(enabled);
CREATE INDEX IF NOT EXISTS idx_project_source_id ON project(source_id);
CREATE INDEX IF NOT EXISTS idx_project_enabled ON project(enabled);
CREATE INDEX IF NOT EXISTS idx_route_rule_project_id ON route_rule(project_id);
CREATE INDEX IF NOT EXISTS idx_route_rule_enabled ON route_rule(enabled);
CREATE INDEX IF NOT EXISTS idx_route_rule_priority ON route_rule(priority);
CREATE INDEX IF NOT EXISTS idx_plugin_registry_name ON plugin_registry(name);