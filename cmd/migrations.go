package cmd

import (
	"context"
	"database/sql"
)

// runMigrations runs the database schema migrations.
func runMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS route (
		id TEXT PRIMARY KEY,
		pattern TEXT NOT NULL,
		identifier TEXT NOT NULL,
		identifier_source TEXT NOT NULL CHECK (identifier_source IN ('path', 'url_param')),
		target_url TEXT NOT NULL,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS header (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		route_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (route_id) REFERENCES route(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS ip_whitelist (
		id TEXT PRIMARY KEY,
		ip_cidr TEXT NOT NULL UNIQUE,
		description TEXT,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS api_key (
		id TEXT PRIMARY KEY,
		key TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		user_id TEXT NOT NULL,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME
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

	CREATE INDEX IF NOT EXISTS idx_route_identifier ON route(identifier);
	CREATE INDEX IF NOT EXISTS idx_route_enabled ON route(enabled);
	CREATE INDEX IF NOT EXISTS idx_header_route_id ON header(route_id);
	CREATE INDEX IF NOT EXISTS idx_whitelist_ip ON ip_whitelist(ip_cidr);
	CREATE INDEX IF NOT EXISTS idx_whitelist_enabled ON ip_whitelist(enabled);
	CREATE INDEX IF NOT EXISTS idx_api_key_key ON api_key(key);
	CREATE INDEX IF NOT EXISTS idx_api_key_enabled ON api_key(enabled);
	`

	_, err := db.ExecContext(context.Background(), schema)
	return err
}