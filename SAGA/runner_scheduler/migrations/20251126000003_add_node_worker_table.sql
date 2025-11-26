-- +goose Up
CREATE TABLE IF NOT EXISTS node_worker (
    id SERIAL PRIMARY KEY,
    node_id INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    worker_id INTEGER NOT NULL REFERENCES worker(id) ON DELETE CASCADE,
    UNIQUE(node_id, worker_id)
);


-- Remove node_id from worker table since we now use junction table
ALTER TABLE worker DROP COLUMN node_id;

-- +goose Down
ALTER TABLE worker ADD COLUMN node_id INTEGER REFERENCES nodes(id) ON DELETE SET NULL;
DROP TABLE IF EXISTS node_worker;

