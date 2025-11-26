-- name: CreateNodeWorker :one
INSERT INTO node_worker (node_id, worker_id) VALUES ($1, $2) RETURNING *;

-- name: GetWorkersByNodeID :many
SELECT w.* FROM worker w
JOIN node_worker nw ON w.id = nw.worker_id
WHERE nw.node_id = $1;

-- name: GetNodeByWorkerID :one
SELECT n.* FROM nodes n
JOIN node_worker nw ON n.id = nw.node_id
WHERE nw.worker_id = $1;

-- name: DeleteNodeWorker :exec
DELETE FROM node_worker WHERE node_id = $1 AND worker_id = $2;

