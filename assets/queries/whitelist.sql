-- IP whitelist queries

-- name: GetWhitelistEntry :one
SELECT * FROM ip_whitelist WHERE id = ?;

-- name: GetWhitelistByIP :one
SELECT * FROM ip_whitelist WHERE ip_cidr = ?;

-- name: ListWhitelist :many
SELECT * FROM ip_whitelist ORDER BY created_at DESC;

-- name: ListEnabledWhitelist :many
SELECT * FROM ip_whitelist WHERE enabled = 1 ORDER BY created_at DESC;

-- name: CreateWhitelistEntry :one
INSERT INTO ip_whitelist (id, ip_cidr, description, enabled)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateWhitelistEntry :one
UPDATE ip_whitelist
SET ip_cidr = ?, description = ?, enabled = ?
WHERE id = ?
RETURNING *;

-- name: DeleteWhitelistEntry :exec
DELETE FROM ip_whitelist WHERE id = ?;