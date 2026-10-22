-- +goose Up
create table public.tariffs (
	plan text primary key,
	rate_minor bigint,
	advance_minor bigint
);
create table public.trips (
	rental_id uuid primary key,
	user_id uuid,
	scooter_id uuid,
	plan text,
	rate_minor bigint,
	advance_minor bigint,
	charged_minor bigint,
	total_minor bigint,
	adjustment_minor bigint,
	billed_minutes bigint,
	status text,
	rental_version bigint,
	reserved_at timestamptz,
	started_at timestamptz,
	finished_at timestamptz,
	version bigint
);

-- +goose Down
drop table public.trips;
drop table public.tariffs;
