-- API key queries

-- name: GetAPIKey :one
SELECT * FROM api_key WHERE id = ?;

-- name: GetAPIKeyByKey :one
SELECT * FROM api_key WHERE key = ?;

-- name: GetAPIKeyByKeyEnabled :one
SELECT * FROM api_key WHERE key = ? AND enabled = 1;

-- name: ListAPIKeys :many
SELECT * FROM api_key ORDER BY created_at DESC;

-- name: CreateAPIKey :one
INSERT INTO api_key (id, key, name, description, user_id, enabled)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateAPIKey :one
UPDATE api_key
SET name = ?, description = ?, user_id = ?, enabled = ?
WHERE id = ?
RETURNING *;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_key SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteAPIKey :exec
DELETE FROM api_key WHERE id = ?;