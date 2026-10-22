-- +goose Up
create table public.scooters (
	id uuid primary key,
	number text unique,
	brand text,
	is_active boolean,
	owner text,
	ridden_seconds bigint,
	version bigint,
	updated_at timestamptz
);

-- +goose Down
drop table public.scooters;
