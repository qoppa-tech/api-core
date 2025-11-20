-- name: CreateClient :one
INSERT INTO clients (
    salon_id,
    name,
    phone,
    email,
    birthday
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetClientByID :one
SELECT * FROM clients
WHERE id = $1 LIMIT 1;

-- name: GetClientByPhone :one
SELECT * FROM clients
WHERE salon_id = $1 AND phone = $2 LIMIT 1;

-- name: ListClientsBySalonID :many
SELECT * FROM clients
WHERE salon_id = $1
ORDER BY name;

-- name: SearchClientsByName :many
SELECT * FROM clients
WHERE salon_id = $1 AND name ILIKE '%' || $2 || '%'
ORDER BY name
LIMIT 50;

-- name: SearchClientsByPhone :many
SELECT * FROM clients
WHERE salon_id = $1 AND phone LIKE '%' || $2 || '%'
ORDER BY name
LIMIT 50;

-- name: UpdateClient :one
UPDATE clients
SET 
    name = COALESCE(sqlc.narg(name), name),
    phone = COALESCE(sqlc.narg(phone), phone),
    email = COALESCE(sqlc.narg(email), email),
    birthday = COALESCE(sqlc.narg(birthday), birthday)
WHERE id = $1
RETURNING *;

-- name: UpdateClientAppointmentStats :exec
UPDATE clients
SET 
    total_appointments = total_appointments + 1,
    last_appointment = $2
WHERE id = $1;

-- name: DeleteClient :exec
DELETE FROM clients
WHERE id = $1;
