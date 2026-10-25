-- +goose Up
CREATE TABLE notes (
   id UUID PRIMARY KEY,
   username VARCHAR(255) NOT NULL,
   title VARCHAR(255) NOT NULL,
   content TEXT NOT NULL,
   updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
   created_at TIMESTAMP WITH TIME ZONE NOT NULL,
   version BIGINT NOT NULL,

   CONSTRAINT fk_notes_user FOREIGN KEY (username)
      REFERENCES users(username) ON DELETE CASCADE
);

CREATE INDEX idx_notes_username ON notes(username);

-- +goose Down
DROP TABLE notes;
