-- +goose Up
CREATE TABLE IF NOT EXISTS nodes (
    id SERIAL PRIMARY KEY,
    address TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS worker (
    id SERIAL PRIMARY KEY,
    camera_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    scenario_uuid UUID NOT NULL,
    state TEXT NOT NULL DEFAULT 'pending',
    node_id INTEGER REFERENCES nodes(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS worker;
DROP TABLE IF EXISTS nodes;

