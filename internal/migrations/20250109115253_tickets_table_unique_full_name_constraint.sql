-- +goose Up
-- +goose StatementBegin
ALTER TABLE tickets
    ADD CONSTRAINT full_name_unique UNIQUE (full_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tickets
    DROP CONSTRAINT full_name_unique;
-- +goose StatementEnd
