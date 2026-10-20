-- +goose Up
CREATE TABLE users (
   username VARCHAR(255) PRIMARY KEY,
   password_hash VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE users;
