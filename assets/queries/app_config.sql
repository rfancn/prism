-- name: GetAppConfig :one
SELECT key, value, value_type, description, updated_at
FROM app_config
WHERE key = ?;

-- name: SetAppConfig :exec
INSERT INTO app_config (key, value, value_type, description, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET
    value = excluded.value,
    value_type = excluded.value_type,
    description = excluded.description,
    updated_at = CURRENT_TIMESTAMP;

-- name: ListAppConfig :many
SELECT key, value, value_type, description, updated_at
FROM app_config
ORDER BY key;

-- name: DeleteAppConfig :exec
DELETE FROM app_config WHERE key = ?;