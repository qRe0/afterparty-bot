-- +goose Up
-- +goose StatementBegin
ALTER TABLE tickets
    ADD COLUMN ticket_price INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tickets
    DROP COLUMN IF EXISTS ticket_price;
-- +goose StatementEnd
