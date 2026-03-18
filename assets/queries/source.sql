-- Source queries

-- name: GetSource :one
SELECT * FROM source WHERE id = ?;

-- name: GetSourceByName :one
SELECT * FROM source WHERE name = ?;

-- name: ListSources :many
SELECT * FROM source ORDER BY created_at DESC;

-- name: CreateSource :one
INSERT INTO source (id, name, description)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateSource :one
UPDATE source
SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteSource :exec
DELETE FROM source WHERE id = ?;