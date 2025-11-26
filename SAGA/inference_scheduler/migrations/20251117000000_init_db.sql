-- +goose Up
-- +goose StatementBegin

-- Таблица inbox_start_scenario для паттерна Inbox в SAGA
CREATE TABLE inbox_start_scenario (
    outbox_uuid UUID NOT NULL PRIMARY KEY,
    camera_id INTEGER NOT NULL,
    scenario_uuid UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'received' CHECK (status IN ('received', 'in_process', 'processed')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);


-- Комментарии к таблице
COMMENT ON TABLE inbox_start_scenario IS 'Inbox pattern table for idempotent message processing in SAGA (start scenario events)';
COMMENT ON COLUMN inbox_start_scenario.outbox_uuid IS 'Unique identifier from the outbox message (serves as primary key for idempotency)';
COMMENT ON COLUMN inbox_start_scenario.camera_id IS 'ID of the camera associated with the scenario';
COMMENT ON COLUMN inbox_start_scenario.scenario_uuid IS 'UUID of the scenario being started';
COMMENT ON COLUMN inbox_start_scenario.status IS 'Processing status: received, in_process, processed';
COMMENT ON COLUMN inbox_start_scenario.created_at IS 'Timestamp when the message was first received';
COMMENT ON COLUMN inbox_start_scenario.updated_at IS 'Timestamp when the message status was last updated';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS inbox_start_scenario;

-- +goose StatementEnd
