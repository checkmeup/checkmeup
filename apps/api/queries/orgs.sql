-- name: GetOrgByID :one
SELECT * FROM orgs WHERE id = $1;
