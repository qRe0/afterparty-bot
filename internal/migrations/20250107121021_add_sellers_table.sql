-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ticket_sellers
(
    ticket_id    INTEGER,
    seller_tag   VARCHAR(255) NOT NULL,
    seller_tg_id VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ticket_sellers;
-- +goose StatementEnd
