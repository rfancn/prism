-- Rollback migration: Drop all tables and indexes
-- 按依赖关系倒序删除（先删除有外键依赖的表）

-- ============================================================
-- 删除索引
-- ============================================================
DROP INDEX IF EXISTS idx_plugin_registry_name;
DROP INDEX IF EXISTS idx_route_rule_priority;
DROP INDEX IF EXISTS idx_route_rule_enabled;
DROP INDEX IF EXISTS idx_route_rule_project_id;
DROP INDEX IF EXISTS idx_project_enabled;
DROP INDEX IF EXISTS idx_project_source_id;
DROP INDEX IF EXISTS idx_source_enabled;
DROP INDEX IF EXISTS idx_whitelist_enabled;
DROP INDEX IF EXISTS idx_whitelist_ip;

-- ============================================================
-- 删除表（按依赖关系倒序）
-- ============================================================

-- route_rule 依赖于 project，先删除
DROP TABLE IF EXISTS route_rule;

-- project 依赖于 source，先删除
DROP TABLE IF EXISTS project;

-- 无外键依赖的表，可按任意顺序删除
DROP TABLE IF EXISTS source;
DROP TABLE IF EXISTS plugin_registry;
DROP TABLE IF EXISTS tls_config;
DROP TABLE IF EXISTS ip_whitelist;