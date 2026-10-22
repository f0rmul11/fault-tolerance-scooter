-- +goose Up
create table public.subscriptions (
	user_id uuid primary key,
	plan text,
	updated_at timestamptz
);

-- +goose Down
drop table public.subscriptions;
