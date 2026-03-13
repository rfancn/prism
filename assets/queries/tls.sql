-- TLS config queries

-- name: GetTLSConfig :one
SELECT * FROM tls_config WHERE id = 1;

-- name: UpsertTLSConfig :exec
INSERT INTO tls_config (id, enabled, cert_file, key_file, auto_cert, domains)
VALUES (1, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    enabled = excluded.enabled,
    cert_file = excluded.cert_file,
    key_file = excluded.key_file,
    auto_cert = excluded.auto_cert,
    domains = excluded.domains,
    updated_at = CURRENT_TIMESTAMP;