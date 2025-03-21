-- name: GetAliveSession :one
SELECT * FROM sessions
WHERE session_id = $1 AND ends_at < CURRENT_TIMESTAMP;

-- name: RefreshSession :one
UPDATE sessions
SET ends_at = CURRENT_TIMESTAMP + MAKE_INTERVAL(secs => @session_duration::bigint), updated_at = CURRENT_TIMESTAMP
WHERE session_id = $1
RETURNING *;

-- name: GetDeviceByExternalDeviceID :one
SELECT * from devices
WHERE external_device_id = $1;

-- name: CreateDevice :one
INSERT INTO devices (device_id, source, external_device_id, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateSession :one
INSERT INTO sessions (session_id, device_id, ends_at)
VALUES ($1, $2, CURRENT_TIMESTAMP + MAKE_INTERVAL(secs => @session_duration::bigint))
RETURNING *;

-- name: RemoveStaleSessions :exec
DELETE FROM sessions
WHERE ends_at <= CURRENT_TIMESTAMP - MAKE_INTERVAL(days => @days);
