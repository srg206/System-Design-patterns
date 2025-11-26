-- name: CreateNode :one
INSERT INTO nodes (address) VALUES ($1) RETURNING *;

-- name: GetNodeByID :one
SELECT * FROM nodes WHERE id = $1;

-- name: GetAllNodes :many
SELECT * FROM nodes;

-- name: DeleteNode :exec
DELETE FROM nodes WHERE id = $1;

