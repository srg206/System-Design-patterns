-- +goose Up
-- +goose StatementBegin

CREATE TABLE node (
    id TEXT PRIMARY KEY,
    addr TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE node_worker
    ADD CONSTRAINT fk_node_worker_node FOREIGN KEY (node_id) REFERENCES node(id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE node_worker DROP CONSTRAINT IF EXISTS fk_node_worker_node;
DROP TABLE IF EXISTS node;

-- +goose StatementEnd

