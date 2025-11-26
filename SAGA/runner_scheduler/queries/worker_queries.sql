-- name: CreateWorker :one
INSERT INTO worker (
    camera_id,
    url,
    scenario_uuid,
    state
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetWorkerByID :one
SELECT * FROM worker WHERE id = $1;

-- name: GetWorkersByScenarioUUID :many
SELECT * FROM worker WHERE scenario_uuid = $1;

-- name: UpdateWorkerState :one
UPDATE worker SET state = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: DeleteWorker :exec
DELETE FROM worker WHERE id = $1;

