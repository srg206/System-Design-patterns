-- name: CreateScenario :one
INSERT INTO scenario (
    uuid,
    camera_id,
    url
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: UpdateScenarioStatusByUUID :exec
UPDATE scenario
SET status = $2,
    updated_at = NOW()
WHERE uuid = $1;

-- name: UpdateScenarioPredictByUUID :exec
UPDATE scenario
SET predict_id = $2,
    updated_at = NOW()
WHERE uuid = $1;

-- name: UpdateScenarioStatusBatch :exec
UPDATE scenario
SET status = $2,
    updated_at = NOW()
WHERE uuid = ANY($1::uuid[]);

-- name: GetScenarioStatusByUUID :one
SELECT uuid, status FROM scenario WHERE uuid = $1;
