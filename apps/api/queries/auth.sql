-- name: CreateOrg :one
INSERT INTO orgs (name, alert_email)
VALUES ($1, $2)
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (org_id, email, password_hash, terms_version, terms_accepted_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1 AND expires_at > NOW();

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token_hash = $1;

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens WHERE user_id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1;

-- name: AcceptUserTerms :one
UPDATE users SET terms_version = $2, terms_accepted_at = NOW(), updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPasswordResetTokenByHash :one
SELECT * FROM password_reset_tokens WHERE token_hash = $1 AND expires_at > NOW();

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_tokens WHERE token_hash = $1;

-- name: DeleteUserPasswordResetTokens :exec
DELETE FROM password_reset_tokens WHERE user_id = $1;
