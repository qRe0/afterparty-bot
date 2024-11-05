-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tickets
(
    id          SERIAL PRIMARY KEY,
    surname     VARCHAR(255) NOT NULL,
    full_name   VARCHAR(255),
    ticket_type VARCHAR(10)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tickets;
-- +goose StatementEnd
