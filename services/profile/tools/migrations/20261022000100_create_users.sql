-- +goose Up
create table public.users (
	id uuid primary key,
	name text,
	email text,
	version bigint,
	updated_at timestamptz
);

-- +goose Down
drop table public.users;
