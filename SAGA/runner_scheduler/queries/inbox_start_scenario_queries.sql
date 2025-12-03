-- name: CreateInboxStartScenario :one
INSERT INTO inbox_start_scenario (
    outbox_uuid,
    camera_id,
    scenario_uuid,
    url
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: DeleteInboxStartScenarios :exec
DELETE FROM inbox_start_scenario
WHERE scenario_uuid = ANY($1::uuid[]);

-- name: GetInboxStartScenariosByScenarioUUIDs :many
SELECT *
FROM inbox_start_scenario
WHERE scenario_uuid = ANY($1::uuid[]);
