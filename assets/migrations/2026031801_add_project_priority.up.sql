-- 项目表添加priority字段
ALTER TABLE project ADD COLUMN priority INTEGER DEFAULT 0;

-- 添加priority索引
CREATE INDEX IF NOT EXISTS idx_project_priority ON project(priority);