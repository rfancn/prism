package cmd

import (
	"context"
	"database/sql"
)

// runMigrations runs the database schema migrations.
func runMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS routes (
		id TEXT PRIMARY KEY,
		pattern TEXT NOT NULL,
		identifier TEXT NOT NULL,
		identifier_source TEXT NOT NULL CHECK (identifier_source IN ('path', 'json_body', 'url_param')),
		target_url TEXT NOT NULL,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS headers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		route_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (route_id) REFERENCES routes(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS ip_whitelist (
		id TEXT PRIMARY KEY,
		ip_cidr TEXT NOT NULL UNIQUE,
		description TEXT,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS api_keys (
		id TEXT PRIMARY KEY,
		key TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		user_id TEXT NOT NULL,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS rate_limits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL UNIQUE,
		requests_per_second INTEGER NOT NULL DEFAULT 100,
		burst INTEGER NOT NULL DEFAULT 200,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS default_rate_limit (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		requests_per_second INTEGER NOT NULL DEFAULT 100,
		burst INTEGER NOT NULL DEFAULT 200
	);

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

	CREATE INDEX IF NOT EXISTS idx_routes_identifier ON routes(identifier);
	CREATE INDEX IF NOT EXISTS idx_routes_enabled ON routes(enabled);
	CREATE INDEX IF NOT EXISTS idx_headers_route_id ON headers(route_id);
	CREATE INDEX IF NOT EXISTS idx_whitelist_ip ON ip_whitelist(ip_cidr);
	CREATE INDEX IF NOT EXISTS idx_whitelist_enabled ON ip_whitelist(enabled);
	CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
	CREATE INDEX IF NOT EXISTS idx_api_keys_enabled ON api_keys(enabled);
	CREATE INDEX IF NOT EXISTS idx_rate_limits_user_id ON rate_limits(user_id);
	`

	_, err := db.ExecContext(context.Background(), schema)
	return err
}