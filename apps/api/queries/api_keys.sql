-- name: CreateAPIKey :one
INSERT INTO api_keys (org_id, key_hash, key_prefix, label)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListAPIKeys :many
SELECT * FROM api_keys WHERE org_id = $1 AND revoked_at IS NULL ORDER BY created_at DESC LIMIT 200;

-- name: CountActiveAPIKeys :one
SELECT COUNT(*) FROM api_keys WHERE org_id = $1 AND revoked_at IS NULL;

-- name: GetActiveAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL;

-- name: RevokeAPIKey :exec
UPDATE api_keys SET revoked_at = NOW() WHERE id = $1 AND org_id = $2 AND revoked_at IS NULL;

-- name: TouchAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = NOW() WHERE id = $1;
