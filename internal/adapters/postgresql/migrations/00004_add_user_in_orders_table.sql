-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
    ADD COLUMN user_id int NOT NULL,
    ADD CONSTRAINT user_id_fk FOREIGN KEY (user_id) REFERENCES users (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
    DROP FOREIGN KEY user_id_fk,
    DROP COLUMN user_id
-- +goose StatementEnd
