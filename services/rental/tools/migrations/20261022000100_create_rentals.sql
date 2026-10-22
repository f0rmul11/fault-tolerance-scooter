-- +goose Up
create table public.rentals (
	id uuid primary key,
	user_id uuid,
	scooter_id uuid,
	idempotency_key uuid,
	request_id uuid,
	scooter_number text,
	scooter_brand text,
	status text,
	reserved_at timestamptz,
	started_at timestamptz,
	finished_at timestamptz,
	updated_at timestamptz,
	version bigint,
	unique (user_id, idempotency_key)
);
create unique index rentals_open_scooter on public.rentals (scooter_id)
	where status in ('RESERVED', 'ACTIVE');
create unique index rentals_open_user on public.rentals (user_id)
	where status in ('RESERVED', 'ACTIVE');

-- +goose Down
drop table public.rentals;
