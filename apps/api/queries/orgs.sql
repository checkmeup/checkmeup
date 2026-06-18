-- name: GetOrgByID :one
SELECT * FROM orgs WHERE id = $1;

-- name: UpdateOrgTelegramChatID :one
UPDATE orgs
SET telegram_chat_id = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrgAlertEmail :one
UPDATE orgs
SET alert_email = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrgEmailAlertsEnabled :one
UPDATE orgs
SET email_alerts_enabled = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
