-- name: GetStaleNodes :many
SELECT
    n.id,
    n.addr,
    n.state,
    MAX(he.updated_at)::timestamp AS last_heartbeat
FROM node n
LEFT JOIN heartbeat_events he ON he.node_id = n.id
GROUP BY n.id, n.addr, n.state, n.created_at
HAVING COALESCE(MAX(he.updated_at), n.created_at) < NOW() - make_interval(secs => $1);

-- name: MarkNodeAsError :exec
UPDATE node SET state = 'error' WHERE id = $1 AND state <> 'error';

-- name: GetStaleWorkers :many
SELECT
    w.id,
    w.camera_id,
    w.status,
    he.updated_at AS last_heartbeat
FROM worker w
LEFT JOIN heartbeat_events he ON he.camera_id = w.camera_id::text
WHERE COALESCE(he.updated_at, w.created_at) < NOW() - make_interval(secs => $1)
  AND w.status <> 'error';

-- name: MarkWorkerAsError :exec
UPDATE worker SET status = 'error', updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND status <> 'error';

-- name: GetWorkersWithoutHeartbeat :many
SELECT
    w.id,
    w.camera_id
FROM worker w
LEFT JOIN heartbeat_events he ON he.camera_id = w.camera_id::text
WHERE he.camera_id IS NULL;


