-- +goose Up
-- +goose StatementBegin

-- Orders table stores purchase order information
-- Таблица scenario
CREATE TABLE scenario (
    uuid UUID NOT NULL UNIQUE,
    camera_id INTEGER NOT NULL,
    predict_id INTEGER,
    status text DEFAULT 'init_startup',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP 
);

-- Таблица outbox
CREATE TABLE outbox_scenario (
    outbox_uuid UUID NOT NULL,
    scenario_uuid UUID NOT NULL REFERENCES scenario(uuid),
    payload JSONB NOT NULL,
    state text DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);




-- Comments for outbox table
COMMENT ON TABLE outbox_scenario IS 'Outbox pattern table for reliable message publishing in SAGA';
COMMENT ON COLUMN outbox_scenario.outbox_uuid IS 'Unique identifier for the outbox message';
COMMENT ON COLUMN outbox_scenario.scenario_uuid IS 'UUID of the associated scenario';
COMMENT ON COLUMN outbox_scenario.payload IS 'JSON data of the message payload';
COMMENT ON COLUMN outbox_scenario.state IS 'State of the message (pending, sent, failed)';
COMMENT ON COLUMN outbox_scenario.created_at IS 'Timestamp when the message was created';
COMMENT ON COLUMN outbox_scenario.updated_at IS 'Timestamp when the message was last updated';

-- Comments for scenario table
COMMENT ON TABLE scenario IS 'Scenario table for storing scenario state and camera prediction';
COMMENT ON COLUMN scenario.uuid IS 'Unique identifier for the scenario (UUID format)';
COMMENT ON COLUMN scenario.camera_id IS 'ID of the camera being used in the scenario';
COMMENT ON COLUMN scenario.predict_id IS 'ID of the predicted person';
COMMENT ON COLUMN scenario.status IS 'Status of the scenario (init_startup, in_startup_processing, active, init_shutdown, in_shutdown_processing, inactive)';
COMMENT ON COLUMN scenario.created_at IS 'Timestamp when the scenario was created';
COMMENT ON COLUMN scenario.updated_at IS 'Timestamp when the scenario was last updated';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS scenario;
DROP TABLE IF EXISTS outbox_scenario;

-- +goose StatementEnd
