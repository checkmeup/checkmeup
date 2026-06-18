-- name: CreateFeatureSuggestion :one
INSERT INTO feature_suggestions (org_id, user_id, text)
VALUES ($1, $2, $3)
RETURNING *;
