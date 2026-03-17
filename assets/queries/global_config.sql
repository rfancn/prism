-- Global config queries

-- name: GetGlobalConfig :one
SELECT * FROM global_config WHERE key = ?;

-- name: SetGlobalConfig :exec
INSERT INTO global_config (key, value, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP;