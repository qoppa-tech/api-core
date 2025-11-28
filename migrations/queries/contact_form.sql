-- name: CreateContactForm :one
INSERT INTO contact_form (
    full_name,
    email,
    phone_number,
    salon_name,
    subject,
    message
  )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: ListContactForms :many
SELECT *
FROM contact_form
ORDER BY created_at DESC OFFSET $1
LIMIT $2;
-- name: ListContactFormsHasNotResponded :many
SELECT *
FROM contact_form
WHERE response IS NULL
ORDER BY created_at DESC OFFSET $1
LIMIT $2;