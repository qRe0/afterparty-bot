-- +goose Up
-- +goose StatementBegin
ALTER TABLE tickets
    ADD COLUMN seller_name VARCHAR(255) DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tickets
    DROP COLUMN IF EXISTS seller_name;
-- +goose StatementEnd
