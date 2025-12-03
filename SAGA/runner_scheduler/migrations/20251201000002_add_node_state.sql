-- +goose Up
-- +goose StatementBegin
ALTER TABLE node ADD COLUMN state TEXT NOT NULL DEFAULT 'up' ;
COMMENT ON COLUMN node.state IS 'Node health state: up (healthy), error (no heartbeat)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE node DROP COLUMN state;
-- +goose StatementEnd

