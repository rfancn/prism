-- Route queries

-- name: GetRoute :one
SELECT * FROM route WHERE id = ?;

-- name: GetRouteByIdentifier :one
SELECT * FROM route WHERE identifier = ? AND enabled = 1;

-- name: ListRoutes :many
SELECT * FROM route ORDER BY created_at DESC;

-- name: ListEnabledRoutes :many
SELECT * FROM route WHERE enabled = 1 ORDER BY created_at DESC;

-- name: CreateRoute :one
INSERT INTO route (id, pattern, identifier, identifier_source, target_url, enabled)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateRoute :one
UPDATE route
SET pattern = ?, identifier = ?, identifier_source = ?, target_url = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM route WHERE id = ?;