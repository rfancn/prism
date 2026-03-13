-- Rate limit queries

-- name: GetRateLimit :one
SELECT * FROM rate_limit WHERE id = ?;

-- name: GetRateLimitByUserID :one
SELECT * FROM rate_limit WHERE user_id = ?;

-- name: ListRateLimits :many
SELECT * FROM rate_limit ORDER BY user_id;

-- name: CreateRateLimit :one
INSERT INTO rate_limit (user_id, requests_per_second, burst)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateRateLimit :one
UPDATE rate_limit
SET requests_per_second = ?, burst = ?, updated_at = CURRENT_TIMESTAMP
WHERE user_id = ?
RETURNING *;

-- name: DeleteRateLimit :exec
DELETE FROM rate_limit WHERE user_id = ?;

-- name: GetDefaultRateLimit :one
SELECT * FROM default_rate_limit WHERE id = 1;

-- name: UpsertDefaultRateLimit :exec
INSERT INTO default_rate_limit (id, requests_per_second, burst)
VALUES (1, ?, ?)
ON CONFLICT(id) DO UPDATE SET requests_per_second = ?, burst = ?;