-- name: CreateService :one
INSERT INTO services (
    salon_id,
    professional_id,
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

-- name: ListServicesByProfessionalID :many
SELECT * FROM services
WHERE professional_id = $1 AND active = true
ORDER BY name;

-- name: ListServicesBySalonAndProfessional :many
SELECT * FROM services
WHERE salon_id = $1 AND (professional_id = $2 OR professional_id IS NULL) AND active = true
ORDER BY name;

-- name: UpdateService :one
UPDATE services
SET 
    name = COALESCE(sqlc.narg(name), name),
    duration = COALESCE(sqlc.narg(duration), duration),
    price = COALESCE(sqlc.narg(price), price),
    active = COALESCE(sqlc.narg(active), active),
    professional_id = COALESCE(sqlc.narg(professional_id), professional_id)
WHERE id = $1
RETURNING *;

-- name: DeleteService :exec
DELETE FROM services
WHERE id = $1;
