-- +goose Up
-- +goose StatementBegin

CREATE TABLE heartbeat_events (
    node_id INTEGER NOT NULL,
    camera_id TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_heartbeat_events_camera_id ON heartbeat_events (camera_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_heartbeat_events_camera_id;
DROP TABLE IF EXISTS heartbeat_events;

-- +goose StatementEnd


