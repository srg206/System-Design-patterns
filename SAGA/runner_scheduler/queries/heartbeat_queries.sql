-- name: UpsertHeartbeat :exec
INSERT INTO heartbeat_events (node_id, camera_id, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (camera_id) DO UPDATE
SET node_id = EXCLUDED.node_id,
    updated_at = EXCLUDED.updated_at
WHERE heartbeat_events.updated_at < EXCLUDED.updated_at;

