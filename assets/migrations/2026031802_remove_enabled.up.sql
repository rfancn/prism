-- 移除各表的enabled字段

-- 1. source表移除enabled字段
CREATE TABLE source_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO source_new SELECT id, name, description, created_at, updated_at FROM source;
DROP TABLE source;
ALTER TABLE source_new RENAME TO source;

-- 2. project表移除enabled字段
CREATE TABLE project_new (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_id) REFERENCES source(id) ON DELETE CASCADE,
    UNIQUE(source_id, name)
);
INSERT INTO project_new SELECT id, source_id, name, description, priority, created_at, updated_at FROM project;
DROP TABLE project;
ALTER TABLE project_new RENAME TO project;

-- 3. route_rule表移除enabled字段
CREATE TABLE route_rule_new (
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
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    UNIQUE(project_id, name)
);
INSERT INTO route_rule_new SELECT id, project_id, name, match_type, path_pattern, cel_expression, plugin_name, target_url, priority, created_at, updated_at FROM route_rule;
DROP TABLE route_rule;
ALTER TABLE route_rule_new RENAME TO route_rule;

-- 4. plugin_registry表移除enabled字段
CREATE TABLE plugin_registry_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    version TEXT,
    command TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO plugin_registry_new SELECT id, name, description, version, command, created_at FROM plugin_registry;
DROP TABLE plugin_registry;
ALTER TABLE plugin_registry_new RENAME TO plugin_registry;

-- 5. 删除已不存在的索引
DROP INDEX IF EXISTS idx_source_enabled;
DROP INDEX IF EXISTS idx_project_enabled;
DROP INDEX IF EXISTS idx_route_rule_enabled;

-- 6. 重建剩余索引
CREATE INDEX IF NOT EXISTS idx_whitelist_ip ON ip_whitelist(ip_cidr);
CREATE INDEX IF NOT EXISTS idx_whitelist_enabled ON ip_whitelist(enabled);
CREATE INDEX IF NOT EXISTS idx_project_source_id ON project(source_id);
CREATE INDEX IF NOT EXISTS idx_project_priority ON project(priority);
CREATE INDEX IF NOT EXISTS idx_route_rule_project_id ON route_rule(project_id);
CREATE INDEX IF NOT EXISTS idx_route_rule_priority ON route_rule(priority);
CREATE INDEX IF NOT EXISTS idx_plugin_registry_name ON plugin_registry(name);