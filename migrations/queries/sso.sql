-- name: CreateSSO :one
INSERT INTO sso (
    user_id,
    provider,
    provider_user_id,
    access_token,
    refresh_token,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetSSOByProvider :one
SELECT * FROM sso
WHERE provider = $1 AND provider_user_id = $2 LIMIT 1;

-- name: GetSSOByUserID :many
SELECT * FROM sso
WHERE user_id = $1;

-- name: UpdateSSOTokens :one
UPDATE sso
SET 
    access_token = $2,
    refresh_token = $3,
    expires_at = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSSO :exec
DELETE FROM sso
WHERE id = $1;

-- name: DeleteSSOByUserID :exec
DELETE FROM sso
WHERE user_id = $1;
