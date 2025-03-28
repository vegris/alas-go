-- name: GetSession :one
SELECT * FROM sessions WHERE session_id = $1;

-- name: GetDevice :one
SELECT * FROM devices WHERE device_id = $1;
