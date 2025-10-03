-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_clients_surname_gin_trgm ON tickets USING gin (surname gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_clients_surname_gin_trgm;
-- +goose StatementEnd
