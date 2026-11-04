-- +goose Up
-- Историческая выборка для чтения скидки, без создания аренд и событий.
insert into public.trips (
    rental_id, user_id, scooter_id, plan, rate_minor, advance_minor,
    charged_minor, total_minor, adjustment_minor, billed_minutes, status,
    rental_version, reserved_at, started_at, finished_at, version
)
select md5('discount-trip-' || user_id::text || '-' || n::text)::uuid,
       user_id,
       '433c55b6-407f-45d5-81ec-66012f917509'::uuid,
       'BASE', 1000, 5000, 1000, 1000, -4000, 1, 'COMPLETED', 3,
       '2026-01-01T00:00:00Z'::timestamptz,
       '2026-01-01T00:00:01Z'::timestamptz,
       '2026-01-01T00:01:01Z'::timestamptz, 3
from (values
    ('704137fe-93b3-4113-8e09-ed1df4d0e0ea'::uuid, 12),
    ('57172605-9d97-4245-9417-8714196b1d20'::uuid, 55)
) as users(user_id, trips)
cross join lateral generate_series(1, trips) as n
on conflict do nothing;

-- Незавершённые записи не должны попадать в счётчик.
insert into public.trips (
    rental_id, user_id, scooter_id, plan, rate_minor, advance_minor,
    charged_minor, status, rental_version, reserved_at, started_at, version
)
select md5('discount-excluded-' || status)::uuid,
       '704137fe-93b3-4113-8e09-ed1df4d0e0ea'::uuid,
       '433c55b6-407f-45d5-81ec-66012f917509'::uuid,
       'BASE', 1000, 5000, case when status = 'ACTIVE' then 5000 else 0 end,
       status, case when status = 'RESERVED' then 1 else 2 end,
       '2026-01-01T00:00:00Z'::timestamptz,
       case when status = 'ACTIVE' then '2026-01-01T00:00:01Z'::timestamptz end,
       case when status = 'RESERVED' then 1 else 2 end
from (values ('RESERVED'), ('ACTIVE'), ('CANCELED')) as states(status)
on conflict do nothing;

-- +goose Down
delete from public.trips
where rental_id in (
    select md5('discount-trip-' || user_id::text || '-' || n::text)::uuid
    from (values
        ('704137fe-93b3-4113-8e09-ed1df4d0e0ea'::uuid, 12),
        ('57172605-9d97-4245-9417-8714196b1d20'::uuid, 55)
    ) as users(user_id, trips)
    cross join lateral generate_series(1, trips) as n
)
or rental_id in (
    select md5('discount-excluded-' || status)::uuid
    from (values ('RESERVED'), ('ACTIVE'), ('CANCELED')) as states(status)
);
