-- Header queries

-- name: GetHeader :one
SELECT * FROM header WHERE id = ?;

-- name: GetHeadersByRouteID :many
SELECT * FROM header WHERE route_id = ? ORDER BY key;

-- name: CreateHeader :one
INSERT INTO header (route_id, key, value)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateHeader :one
UPDATE header SET key = ?, value = ? WHERE id = ?
RETURNING *;

-- name: DeleteHeader :exec
DELETE FROM header WHERE id = ?;

-- name: DeleteHeadersByRouteID :exec
DELETE FROM header WHERE route_id = ?;