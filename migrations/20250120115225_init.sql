-- +goose Up
-- +goose StatementBegin
create table if not exists bcmon.owner (
    id serial primary key,
    address varchar(255) unique not null
);

create table if not exists bcmon.contract (
    id serial primary key,
    address varchar(255) unique not null
);

create table if not exists bcmon.nft (
    owner_id integer references owner(id),
    contract_id integer references contract(id),
    count integer not null default 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
