-- +goose Up
-- +goose StatementBegin

CREATE TABLE worker (
    id SERIAL PRIMARY KEY,
    camera_id INTEGER NOT NULL UNIQUE,
    scenario_uuid UUID NOT NULL,
    url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE node_worker (
    id SERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    worker_id INTEGER NOT NULL REFERENCES worker(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(node_id, worker_id)
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS node_worker;
DROP TABLE IF EXISTS worker;

-- +goose StatementEnd

