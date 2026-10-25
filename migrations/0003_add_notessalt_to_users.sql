-- +goose Up
ALTER TABLE users ADD COLUMN notes_salt BYTEA;

-- +goose Down
ALTER TABLE users DROP COLUMN notes_salt;
