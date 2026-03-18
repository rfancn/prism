-- 将target_url从project移回route_rule

-- 1. route_rule表添加target_url字段
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
    target_url TEXT NOT NULL DEFAULT '',
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    UNIQUE(project_id, name)
);

-- 2. 从project迁回target_url
INSERT INTO route_rule_new
SELECT rr.id, rr.project_id, rr.name, rr.match_type, rr.path_pattern, rr.cel_expression, rr.plugin_name,
       COALESCE(p.target_url, '') as target_url, rr.priority, rr.created_at, rr.updated_at
FROM route_rule rr
LEFT JOIN project p ON rr.project_id = p.id;

DROP TABLE route_rule;
ALTER TABLE route_rule_new RENAME TO route_rule;

-- 3. 重建索引
CREATE INDEX IF NOT EXISTS idx_route_rule_project_id ON route_rule(project_id);
CREATE INDEX IF NOT EXISTS idx_route_rule_priority ON route_rule(priority);

-- 4. 删除project的target_url字段
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

-- 5. 重建project索引
CREATE INDEX IF NOT EXISTS idx_project_source_id ON project(source_id);
CREATE INDEX IF NOT EXISTS idx_project_priority ON project(priority);