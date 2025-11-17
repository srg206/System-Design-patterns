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

-- name: LockOutboxScenario :exec
UPDATE outbox_scenario
SET locked_until = $2,
    updated_at = NOW()
WHERE outbox_uuid = $1;

-- name: GetPendingOutboxScenarios :many
SELECT * FROM outbox_scenario
WHERE state = 'pending'
  AND (locked_until IS NULL OR locked_until <= NOW())
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: LockOutboxScenariosBatch :exec
UPDATE outbox_scenario
SET locked_until = NOW() + INTERVAL '1 minute',
    updated_at = NOW()
WHERE outbox_uuid = ANY($1::uuid[]);

-- name: MarkOutboxScenariosAsSentBatch :exec
UPDATE outbox_scenario
SET state = 'sent',
    locked_until = NULL,
    updated_at = NOW()
WHERE outbox_uuid = ANY($1::uuid[]);

