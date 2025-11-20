-- name: CreateNotification :one
INSERT INTO notifications (
    appointment_id,
    type,
    scheduled_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetNotificationByID :one
SELECT * FROM notifications
WHERE id = $1 LIMIT 1;

-- name: ListNotificationsByAppointmentID :many
SELECT * FROM notifications
WHERE appointment_id = $1
ORDER BY scheduled_at;

-- name: ListPendingNotifications :many
SELECT * FROM notifications
WHERE status = 'queued' AND scheduled_at <= NOW()
ORDER BY scheduled_at
LIMIT $1;

-- name: ListNotificationsByStatus :many
SELECT * FROM notifications
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: UpdateNotificationStatus :one
UPDATE notifications
SET 
    status = $2,
    sent_at = CASE WHEN $2 = 'sent' THEN NOW() ELSE sent_at END
WHERE id = $1
RETURNING *;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1;

-- name: DeleteNotificationsByAppointmentID :exec
DELETE FROM notifications
WHERE appointment_id = $1;
