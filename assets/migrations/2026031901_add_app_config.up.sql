-- 添加应用配置表
CREATE TABLE IF NOT EXISTS app_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL DEFAULT 'string',
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);