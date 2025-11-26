-- +goose Up
ALTER TABLE nodes ADD COLUMN state TEXT NOT NULL DEFAULT 'active';

-- +goose Down
ALTER TABLE nodes DROP COLUMN state;

