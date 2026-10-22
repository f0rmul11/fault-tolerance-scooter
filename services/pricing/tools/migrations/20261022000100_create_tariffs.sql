-- +goose Up
create table public.tariffs (
	plan text primary key,
	rate_minor bigint,
	advance_minor bigint
);

-- +goose Down
drop table public.tariffs;
