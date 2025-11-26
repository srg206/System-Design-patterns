-- +goose Up
-- +goose StatementBegin

ALTER TABLE inbox_start_scenario ADD COLUMN url TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN inbox_start_scenario.url IS 'URL to connect to camera';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE inbox_start_scenario DROP COLUMN IF EXISTS url;

-- +goose StatementEnd

