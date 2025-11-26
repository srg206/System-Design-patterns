-- +goose Up
-- +goose StatementBegin

ALTER TABLE scenario ADD COLUMN url TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN scenario.url IS 'URL to connect to camera';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE scenario DROP COLUMN IF EXISTS url;

-- +goose StatementEnd

