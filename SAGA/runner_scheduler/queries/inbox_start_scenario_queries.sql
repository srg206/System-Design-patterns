-- name: CreateInboxStartScenario :one
INSERT INTO inbox_start_scenario (
    outbox_uuid,
    camera_id,
    scenario_uuid,
    url
) VALUES (
    $1, $2, $3, $4
) RETURNING *;
