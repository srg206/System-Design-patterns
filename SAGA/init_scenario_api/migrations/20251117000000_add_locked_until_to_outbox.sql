-- +goose Up
-- +goose StatementBegin

-- Add locked_until column to outbox_scenario table
ALTER TABLE outbox_scenario 
ADD COLUMN locked_until TIMESTAMP DEFAULT NULL;

-- Add comment for the new column
COMMENT ON COLUMN outbox_scenario.locked_until IS 'Timestamp until which the outbox message is locked from being processed';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove locked_until column from outbox_scenario table
ALTER TABLE outbox_scenario 
DROP COLUMN IF EXISTS locked_until;

-- +goose StatementEnd

