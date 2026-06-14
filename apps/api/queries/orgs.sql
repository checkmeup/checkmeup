-- name: GetOrgByID :one
SELECT * FROM orgs WHERE id = $1;

-- name: UpdateOrgTelegramChatID :one
UPDATE orgs
SET telegram_chat_id = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
