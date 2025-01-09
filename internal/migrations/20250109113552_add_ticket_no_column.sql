-- +goose Up
-- +goose StatementBegin
ALTER TABLE tickets
    ADD COLUMN ticketNo INTEGER UNIQUE ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tickets
    DROP COLUMN IF EXISTS ticketNo;
-- +goose StatementEnd
