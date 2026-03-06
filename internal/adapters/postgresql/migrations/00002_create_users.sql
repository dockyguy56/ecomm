-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
   id int PRIMARY KEY NOT NULL GENERATED ALWAYS AS IDENTITY,
   name varchar(255) NOT NULL,
   email varchar(255) NOT NULL,
   password varchar(255) NOT NULL,
   is_admin bool NOT NULL DEFAULT false,
   created_at TIMESTAMPTZ DEFAULT (now()),
   updated_at TIMESTAMPTZ,
   UNIQUE(email)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users
-- +goose StatementEnd
