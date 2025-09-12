-- +goose Up
alter table users add hashed_password text;

-- +goose Down
alter table users drop hashed_password;