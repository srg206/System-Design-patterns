-- Scenario table stores camera prediction scenarios
CREATE TABLE IF NOT EXISTS scenario (
    uuid UUID NOT NULL UNIQUE,
    camera_id INTEGER NOT NULL,
    url TEXT NOT NULL DEFAULT '',
    predict_id INTEGER,
    status TEXT DEFAULT 'init_startup',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

COMMENT ON TABLE scenario IS 'Scenario table for storing scenario state and camera prediction';
COMMENT ON COLUMN scenario.uuid IS 'Unique identifier for the scenario (UUID format)';
COMMENT ON COLUMN scenario.camera_id IS 'ID of the camera being used in the scenario';
COMMENT ON COLUMN scenario.url IS 'URL to connect to camera';
COMMENT ON COLUMN scenario.predict_id IS 'ID of the predicted person';
COMMENT ON COLUMN scenario.status IS 'Status of the scenario (init_startup, in_startup_processing, active, init_shutdown, in_shutdown_processing, inactive)';
COMMENT ON COLUMN scenario.created_at IS 'Timestamp when the scenario was created';
COMMENT ON COLUMN scenario.updated_at IS 'Timestamp when the scenario was last updated';

-- Outbox table for scenario related events
CREATE TABLE IF NOT EXISTS outbox_scenario (
    outbox_uuid UUID NOT NULL,
    scenario_uuid UUID NOT NULL REFERENCES scenario(uuid),
    payload JSONB NOT NULL,
    state TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    locked_until TIMESTAMP DEFAULT NULL
);

COMMENT ON TABLE outbox_scenario IS 'Outbox pattern table for reliable message publishing in SAGA';
COMMENT ON COLUMN outbox_scenario.scenario_uuid IS 'UUID of the associated scenario from scenario table';
COMMENT ON COLUMN outbox_scenario.payload IS 'JSON data of the message payload';
COMMENT ON COLUMN outbox_scenario.state IS 'State of the message (pending, sent, failed)';
COMMENT ON COLUMN outbox_scenario.created_at IS 'Timestamp when the message was created';
COMMENT ON COLUMN outbox_scenario.updated_at IS 'Timestamp when the message was last updated';
COMMENT ON COLUMN outbox_scenario.locked_until IS 'Timestamp until which the outbox message is locked from being processed';
