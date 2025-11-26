-- name: CreateScenario :one
INSERT INTO scenario (
    uuid,
    camera_id
) VALUES (
    $1, $2
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

-- name: CreateOutboxScenario :one
INSERT INTO outbox_scenario (
    outbox_uuid,
    scenario_uuid,
    payload
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: UpdateOutboxScenarioState :exec
UPDATE outbox_scenario
SET state = $2,
    updated_at = NOW()
WHERE scenario_uuid = $1;
