-- Source queries

-- name: GetSource :one
SELECT * FROM source WHERE id = ?;

-- name: GetSourceByName :one
SELECT * FROM source WHERE name = ?;

-- name: ListSources :many
SELECT * FROM source ORDER BY created_at DESC;

-- name: ListEnabledSources :many
SELECT * FROM source WHERE enabled = 1 ORDER BY created_at DESC;

-- name: CreateSource :one
INSERT INTO source (id, name, description, enabled)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateSource :one
UPDATE source
SET name = ?, description = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteSource :exec
DELETE FROM source WHERE id = ?;