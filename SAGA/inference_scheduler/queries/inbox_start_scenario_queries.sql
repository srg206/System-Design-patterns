-- name: CreateInboxStartScenario :one
INSERT INTO inbox_start_scenario (
    outbox_uuid,
    camera_id,
    scenario_uuid
) VALUES (
    $1, $2, $3
) RETURNING *;
