-- Plugin registry queries

-- name: GetPlugin :one
SELECT * FROM plugin_registry WHERE id = ?;

-- name: GetPluginByName :one
SELECT * FROM plugin_registry WHERE name = ?;

-- name: ListPlugins :many
SELECT * FROM plugin_registry ORDER BY created_at DESC;

-- name: CreatePlugin :one
INSERT INTO plugin_registry (id, name, description, version, command)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdatePlugin :one
UPDATE plugin_registry
SET name = ?, description = ?, version = ?, command = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlugin :exec
DELETE FROM plugin_registry WHERE id = ?;