-- +goose Up
-- +goose StatementBegin
ALTER TABLE worker DROP CONSTRAINT IF EXISTS worker_status_check;
ALTER TABLE worker ADD CONSTRAINT worker_status_check CHECK (status IN ('pending', 'running', 'stopped', 'failed', 'error'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE worker DROP CONSTRAINT IF EXISTS worker_status_check;
ALTER TABLE worker ADD CONSTRAINT worker_status_check CHECK (status IN ('pending', 'running', 'stopped', 'failed'));
-- +goose StatementEnd

