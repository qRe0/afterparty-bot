-- +goose Up
-- +goose StatementBegin
ALTER TABLE tickets
    ADD COLUMN passed_control_zone BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tickets
    DROP COLUMN IF EXISTS passed_control_zone;
-- +goose StatementEnd
