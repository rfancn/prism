-- 删除priority索引和字段
DROP INDEX IF EXISTS idx_project_priority;
-- SQLite不支持DROP COLUMN，需要重建表
CREATE TABLE project_new (
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
INSERT INTO project_new SELECT id, source_id, name, description, enabled, created_at, updated_at FROM project;
DROP TABLE project;
ALTER TABLE project_new RENAME TO project;