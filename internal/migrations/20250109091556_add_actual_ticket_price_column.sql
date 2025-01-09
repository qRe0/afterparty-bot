-- +goose Up
-- +goose StatementBegin
ALTER TABLE tickets
    ADD COLUMN actual_ticket_price INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tickets
    DROP COLUMN IF EXISTS actual_ticket_price;
-- +goose StatementEnd
