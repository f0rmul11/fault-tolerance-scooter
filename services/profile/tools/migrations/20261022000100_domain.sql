-- +goose Up
create table public.users (
	id uuid primary key,
	name text,
	email text,
	version bigint,
	updated_at timestamptz
);
create table public.subscriptions (
	user_id uuid primary key,
	plan text,
	updated_at timestamptz
);

-- +goose Down
drop table public.subscriptions;
drop table public.users;
