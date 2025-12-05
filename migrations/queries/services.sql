-- name: CreateService :one
INSERT INTO services (
    salon_id,
    user_id,
    name,
    duration,
    price,
    active
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetServiceByID :one
SELECT * FROM services
WHERE id = $1 LIMIT 1;

-- name: ListServicesBySalonID :many
SELECT * FROM services
WHERE salon_id = $1
ORDER BY name;

-- name: ListActiveServicesBySalonID :many
SELECT * FROM services
WHERE salon_id = $1 AND active = true
ORDER BY name;

-- name: ListServicesByUserID :many
SELECT * FROM services
WHERE user_id = $1 AND active = true
ORDER BY name;

-- name: ListServicesBySalonAndUser :many
SELECT * FROM services
WHERE salon_id = $1 AND (user_id = $2 OR user_id IS NULL) AND active = true
ORDER BY name;

-- name: UpdateService :one
UPDATE services
SET 
    name = COALESCE(sqlc.narg(name), name),
    duration = COALESCE(sqlc.narg(duration), duration),
    price = COALESCE(sqlc.narg(price), price),
    active = COALESCE(sqlc.narg(active), active),
    user_id = COALESCE(sqlc.narg(user_id), user_id)
WHERE id = $1
RETURNING *;

-- name: DeleteService :exec
DELETE FROM services
WHERE id = $1;
