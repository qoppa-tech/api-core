-- name: CreateSSO :one
INSERT INTO sso (
    user_id,
    provider,
    provider_user_id
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetSSOByProvider :one
SELECT * FROM sso
WHERE provider = $1 AND provider_user_id = $2 LIMIT 1;

-- name: GetSSOByUserID :many
SELECT * FROM sso
WHERE user_id = $1;

-- name: DeleteSSO :exec
DELETE FROM sso
WHERE id = $1;

-- name: DeleteSSOByUserID :exec
DELETE FROM sso
WHERE user_id = $1;
