-- Plugin registry queries

-- name: GetPlugin :one
SELECT * FROM plugin_registry WHERE id = ?;

-- name: GetPluginByName :one
SELECT * FROM plugin_registry WHERE name = ?;

-- name: ListPlugins :many
SELECT * FROM plugin_registry ORDER BY created_at DESC;

-- name: ListEnabledPlugins :many
SELECT * FROM plugin_registry WHERE enabled = 1 ORDER BY created_at DESC;

-- name: CreatePlugin :one
INSERT INTO plugin_registry (id, name, description, version, command, enabled)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdatePlugin :one
UPDATE plugin_registry
SET name = ?, description = ?, version = ?, command = ?, enabled = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlugin :exec
DELETE FROM plugin_registry WHERE id = ?;