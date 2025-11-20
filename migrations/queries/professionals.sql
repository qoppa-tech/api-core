-- name: CreateProfessional :one
INSERT INTO professionals (
    user_id,
    salon_id,
    name,
    photo_url,
    active
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetProfessionalByID :one
SELECT * FROM professionals
WHERE id = $1 LIMIT 1;

-- name: GetProfessionalByUserAndSalon :one
SELECT * FROM professionals
WHERE user_id = $1 AND salon_id = $2 LIMIT 1;

-- name: ListProfessionalsBySalonID :many
SELECT * FROM professionals
WHERE salon_id = $1
ORDER BY name;

-- name: ListActiveProfessionalsBySalonID :many
SELECT * FROM professionals
WHERE salon_id = $1 AND active = true
ORDER BY name;

-- name: UpdateProfessional :one
UPDATE professionals
SET 
    name = COALESCE(sqlc.narg(name), name),
    photo_url = COALESCE(sqlc.narg(photo_url), photo_url),
    active = COALESCE(sqlc.narg(active), active),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProfessional :exec
DELETE FROM professionals
WHERE id = $1;
