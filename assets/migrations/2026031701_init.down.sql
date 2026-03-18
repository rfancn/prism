-- 回滚初始化schema
DROP INDEX IF EXISTS idx_plugin_registry_name;
DROP INDEX IF EXISTS idx_route_rule_priority;
DROP INDEX IF EXISTS idx_route_rule_enabled;
DROP INDEX IF EXISTS idx_route_rule_project_id;
DROP INDEX IF EXISTS idx_project_enabled;
DROP INDEX IF EXISTS idx_project_source_id;
DROP INDEX IF EXISTS idx_source_enabled;
DROP INDEX IF EXISTS idx_whitelist_enabled;
DROP INDEX IF EXISTS idx_whitelist_ip;

DROP TABLE IF EXISTS plugin_registry;
DROP TABLE IF EXISTS route_rule;
DROP TABLE IF EXISTS project;
DROP TABLE IF EXISTS source;
DROP TABLE IF EXISTS tls_config;
DROP TABLE IF EXISTS ip_whitelist;