-- name: CreateUser :one
INSERT INTO users (
    email, 
    password_hash, 
    name, 
    phone, 
    role, 
    salon_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: ListUsersBySalonID :many
SELECT * FROM users
WHERE salon_id = $1
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET 
    email = COALESCE(sqlc.narg(email), email),
    name = COALESCE(sqlc.narg(name), name),
    phone = COALESCE(sqlc.narg(phone), phone),
    role = COALESCE(sqlc.narg(role), role),
    salon_id = COALESCE(sqlc.narg(salon_id), salon_id),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
