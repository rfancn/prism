-- Schema for Prism HTTP/HTTPS request relay tool
-- SQLite database schema

-- Route table: stores routing configurations
CREATE TABLE IF NOT EXISTS route (
    id TEXT PRIMARY KEY,
    pattern TEXT NOT NULL,                    -- Path pattern, e.g., /api/{tenant}/users
    identifier TEXT NOT NULL,                  -- Field name to extract for routing
    identifier_source TEXT NOT NULL CHECK (identifier_source IN ('path', 'url_param')),
    target_url TEXT NOT NULL,                  -- Target URL to forward to
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Header table: stores custom headers for routes
CREATE TABLE IF NOT EXISTS header (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    route_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (route_id) REFERENCES route(id) ON DELETE CASCADE
);

-- IP whitelist table: stores allowed IP addresses/CIDR ranges
CREATE TABLE IF NOT EXISTS ip_whitelist (
    id TEXT PRIMARY KEY,
    ip_cidr TEXT NOT NULL UNIQUE,              -- IP address or CIDR range
    description TEXT,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- API key table: stores API keys for authentication
CREATE TABLE IF NOT EXISTS api_key (
    id TEXT PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    user_id TEXT NOT NULL,                     -- User identifier for rate limiting
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME
);

-- TLS configuration
CREATE TABLE IF NOT EXISTS tls_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER DEFAULT 0,
    cert_file TEXT,
    key_file TEXT,
    auto_cert INTEGER DEFAULT 0,
    domains TEXT,                              -- Comma-separated domains for auto-cert
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_route_identifier ON route(identifier);
CREATE INDEX IF NOT EXISTS idx_route_enabled ON route(enabled);
CREATE INDEX IF NOT EXISTS idx_header_route_id ON header(route_id);
CREATE INDEX IF NOT EXISTS idx_whitelist_ip ON ip_whitelist(ip_cidr);
CREATE INDEX IF NOT EXISTS idx_whitelist_enabled ON ip_whitelist(enabled);
CREATE INDEX IF NOT EXISTS idx_api_key_key ON api_key(key);
CREATE INDEX IF NOT EXISTS idx_api_key_enabled ON api_key(enabled);