-- +goose Up
-- +goose StatementBegin

-- Remove foreign key constraint
ALTER TABLE node_worker DROP CONSTRAINT IF EXISTS fk_node_worker_node;

-- Drop all data and change node_id type in node_worker
TRUNCATE TABLE node_worker;

-- Change node_id type from TEXT to INTEGER in node_worker
ALTER TABLE node_worker ALTER COLUMN node_id TYPE INTEGER USING node_id::INTEGER;

-- Drop and recreate the node table with correct type
DROP TABLE IF EXISTS node CASCADE;

CREATE TABLE node (
    id INTEGER PRIMARY KEY,
    addr TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Recreate foreign key constraint
ALTER TABLE node_worker
    ADD CONSTRAINT fk_node_worker_node FOREIGN KEY (node_id) REFERENCES node(id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE node_worker DROP CONSTRAINT IF EXISTS fk_node_worker_node;

DROP TABLE IF EXISTS node;

CREATE TABLE node (
    id TEXT PRIMARY KEY,
    addr TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE node_worker
    ADD CONSTRAINT fk_node_worker_node FOREIGN KEY (node_id) REFERENCES node(id) ON DELETE CASCADE;

-- +goose StatementEnd

