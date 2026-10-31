-- +goose Up
ALTER TABLE notes ADD COLUMN position BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE notes DROP COLUMN position;
