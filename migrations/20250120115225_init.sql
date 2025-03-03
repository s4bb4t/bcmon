-- +goose Up
-- +goose StatementBegin
create table if not exists public.contract (
    address varchar(255) unique not null
);
-- +goose StatementEnd