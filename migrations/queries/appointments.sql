-- name: CreateAppointment :one
INSERT INTO appointments (
    salon_id,
    professional_id,
    service_id,
    client_name,
    client_phone,
    client_email,
    date,
    start_time,
    end_time,
    status,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetAppointmentByID :one
SELECT * FROM appointments
WHERE id = $1 LIMIT 1;

-- name: ListAppointmentsBySalonID :many
SELECT * FROM appointments
WHERE salon_id = $1
ORDER BY date DESC, start_time DESC;

-- name: ListAppointmentsByDate :many
SELECT * FROM appointments
WHERE salon_id = $1 AND date = $2
ORDER BY start_time;

-- name: ListAppointmentsByDateRange :many
SELECT * FROM appointments
WHERE salon_id = $1 AND date BETWEEN $2 AND $3
ORDER BY date, start_time;

-- name: ListAppointmentsByProfessional :many
SELECT * FROM appointments
WHERE professional_id = $1 AND date = $2
ORDER BY start_time;

-- name: ListAppointmentsByProfessionalAndDateRange :many
SELECT * FROM appointments
WHERE professional_id = $1 AND date BETWEEN $2 AND $3
ORDER BY date, start_time;

-- name: ListAppointmentsByStatus :many
SELECT * FROM appointments
WHERE salon_id = $1 AND status = $2
ORDER BY date DESC, start_time DESC;

-- name: ListAppointmentsByClientPhone :many
SELECT * FROM appointments
WHERE salon_id = $1 AND client_phone = $2
ORDER BY date DESC, start_time DESC;

-- name: UpdateAppointmentStatus :one
UPDATE appointments
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateAppointment :one
UPDATE appointments
SET 
    date = COALESCE(sqlc.narg(date), date),
    start_time = COALESCE(sqlc.narg(start_time), start_time),
    end_time = COALESCE(sqlc.narg(end_time), end_time),
    status = COALESCE(sqlc.narg(status), status),
    notes = COALESCE(sqlc.narg(notes), notes)
WHERE id = $1
RETURNING *;

-- name: DeleteAppointment :exec
DELETE FROM appointments
WHERE id = $1;

-- name: CountAppointmentsByProfessionalAndDate :one
SELECT COUNT(*) FROM appointments
WHERE professional_id = $1 AND date = $2 AND status NOT IN ('cancelled', 'no-show');
