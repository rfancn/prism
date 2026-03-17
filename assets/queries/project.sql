-- Project queries

-- name: GetProject :one
SELECT * FROM project WHERE id = ?;

-- name: GetProjectBySourceAndName :one
SELECT * FROM project WHERE source_id = ? AND name = ?;

-- name: ListProjects :many
SELECT * FROM project ORDER BY created_at DESC;

-- name: ListEnabledProjects :many
SELECT * FROM project WHERE enabled = 1 ORDER BY created_at DESC;

-- name: ListProjectsBySourceID :many
SELECT * FROM project WHERE source_id = ? ORDER BY created_at DESC;

-- name: ListEnabledProjectsBySourceID :many
SELECT * FROM project WHERE source_id = ? AND enabled = 1 ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO project (id, source_id, name, description, enabled)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateProject :one
UPDATE project
SET name = ?, description = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM project WHERE id = ?;