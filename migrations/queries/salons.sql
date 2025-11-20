-- name: CreateSalon :one
INSERT INTO salons (
    name,
    slug,
    owner_id,
    address,
    whatsapp,
    business_hours,
    logo_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetSalonByID :one
SELECT * FROM salons
WHERE id = $1 LIMIT 1;

-- name: GetSalonBySlug :one
SELECT * FROM salons
WHERE slug = $1 LIMIT 1;

-- name: GetSalonByOwnerID :one
SELECT * FROM salons
WHERE owner_id = $1 LIMIT 1;

-- name: ListSalons :many
SELECT * FROM salons
ORDER BY created_at DESC;

-- name: UpdateSalon :one
UPDATE salons
SET 
    name = COALESCE(sqlc.narg(name), name),
    slug = COALESCE(sqlc.narg(slug), slug),
    address = COALESCE(sqlc.narg(address), address),
    whatsapp = COALESCE(sqlc.narg(whatsapp), whatsapp),
    business_hours = COALESCE(sqlc.narg(business_hours), business_hours),
    logo_url = COALESCE(sqlc.narg(logo_url), logo_url),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSalon :exec
DELETE FROM salons
WHERE id = $1;
