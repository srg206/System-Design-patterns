-- name: CreateWorker :one
INSERT INTO worker (
    camera_id,
    scenario_uuid,
    url,
    status
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: CreateNodeWorker :one
INSERT INTO node_worker (
    node_id,
    worker_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetWorkerByCameraID :one
SELECT * FROM worker WHERE camera_id = $1;

-- name: UpdateWorkerStatus :one
UPDATE worker SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 RETURNING *;

-- name: DeleteWorker :exec
DELETE FROM worker WHERE id = $1;

-- name: DeleteNodeWorkerByWorkerID :exec
DELETE FROM node_worker WHERE worker_id = $1;

-- name: GetOldestWorkersByStatus :many
SELECT * FROM worker WHERE status = $1 ORDER BY created_at ASC LIMIT $2;

-- name: GetLeastLoadedNodes :many
SELECT n.id as node_id, n.addr, COUNT(nw.worker_id) as worker_count
FROM node n
LEFT JOIN node_worker nw ON n.id = nw.node_id
GROUP BY n.id, n.addr
ORDER BY worker_count ASC
LIMIT $1;

