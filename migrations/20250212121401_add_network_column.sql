-- +goose Up
-- +goose StatementBegin
alter table public.contract add column network varchar(255);
-- +goose StatementEnd
