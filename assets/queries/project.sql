-- Project queries

-- name: GetProject :one
SELECT * FROM project WHERE id = ?;

-- name: GetProjectBySourceAndName :one
SELECT * FROM project WHERE source_id = ? AND name = ?;

-- name: ListProjects :many
SELECT * FROM project ORDER BY source_id, priority ASC;

-- name: ListProjectsBySourceID :many
SELECT * FROM project WHERE source_id = ? ORDER BY priority ASC;

-- name: CreateProject :one
INSERT INTO project (id, source_id, name, description, target_url, priority)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateProject :one
UPDATE project
SET source_id = ?, name = ?, description = ?, target_url = ?, priority = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM project WHERE id = ?;