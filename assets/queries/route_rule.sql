-- Route rule queries

-- name: GetRouteRule :one
SELECT * FROM route_rule WHERE id = ?;

-- name: ListRouteRules :many
SELECT * FROM route_rule ORDER BY priority ASC, created_at DESC;

-- name: ListRouteRulesByProjectID :many
SELECT * FROM route_rule WHERE project_id = ? ORDER BY priority ASC, created_at DESC;

-- name: CreateRouteRule :one
INSERT INTO route_rule (id, project_id, name, match_type, path_pattern, cel_expression, plugin_name, priority)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateRouteRule :one
UPDATE route_rule
SET name = ?, match_type = ?, path_pattern = ?, cel_expression = ?, plugin_name = ?, priority = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteRouteRule :exec
DELETE FROM route_rule WHERE id = ?;

-- name: DeleteRouteRulesByProjectID :exec
DELETE FROM route_rule WHERE project_id = ?;