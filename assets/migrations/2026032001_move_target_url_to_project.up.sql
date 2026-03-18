-- 移动target_url从route_rule到project

-- 1. 项目表添加target_url字段
ALTER TABLE project ADD COLUMN target_url TEXT;

-- 2. 将route_rule中的target_url迁移到project（取第一个规则的target_url）
UPDATE project SET target_url = (
    SELECT target_url FROM route_rule WHERE route_rule.project_id = project.id LIMIT 1
);

-- 3. 删除route_rule的target_url字段（SQLite需要重建表）
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
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    UNIQUE(project_id, name)
);
INSERT INTO route_rule_new SELECT id, project_id, name, match_type, path_pattern, cel_expression, plugin_name, priority, created_at, updated_at FROM route_rule;
DROP TABLE route_rule;
ALTER TABLE route_rule_new RENAME TO route_rule;

-- 4. 重建索引
CREATE INDEX IF NOT EXISTS idx_route_rule_project_id ON route_rule(project_id);
CREATE INDEX IF NOT EXISTS idx_route_rule_priority ON route_rule(priority);