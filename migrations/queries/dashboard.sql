-- name: DashboardAppointmentStatus :many
SELECT status, COUNT(*) AS count
FROM appointments
WHERE salon_id = $1 AND date >= $2 AND date < $3
GROUP BY status;

-- name: DashboardServiceCounts :many
SELECT a.service_id, COALESCE(s.name, '') AS name, COUNT(*) AS count
FROM appointments a
LEFT JOIN services s ON s.id = a.service_id
WHERE a.salon_id = $1 AND a.date >= $2 AND a.date < $3
GROUP BY a.service_id, s.name
ORDER BY count DESC;

-- name: DashboardNewClients :one
SELECT COUNT(*)
FROM clients
WHERE salon_id = $1 AND created_at >= $2 AND created_at < $3;

-- name: DashboardUniqueClients :one
SELECT COUNT(DISTINCT client_phone)
FROM appointments
WHERE salon_id = $1 AND date >= $2 AND date < $3;

-- name: DashboardActiveClients :one
SELECT COUNT(DISTINCT client_phone)
FROM appointments
WHERE salon_id = $1 AND date >= $2 AND date < $3;

-- name: DashboardEngagedClients :one
SELECT COUNT(*)
FROM (
    SELECT client_phone
    FROM appointments
    WHERE salon_id = $1 AND date >= $2 AND date < $3
    GROUP BY client_phone
    HAVING COUNT(*) >= 2
) t;
