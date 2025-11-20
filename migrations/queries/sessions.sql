-- name: CreateSession :one
INSERT INTO sessions (
    user_id,
    token,
    expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE token = $1 AND expires_at > NOW() LIMIT 1;

-- name: GetSessionByID :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1;

-- name: ListSessionsByUserID :many
SELECT * FROM sessions
WHERE user_id = $1 AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteSessionByToken :exec
DELETE FROM sessions
WHERE token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= NOW();

-- name: DeleteUserSessions :exec
DELETE FROM sessions
WHERE user_id = $1;
