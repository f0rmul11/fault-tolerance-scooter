-- +goose Up
-- Учебные пользователи программы скидок.
insert into public.users (id, name, email, version, updated_at)
values
    ('e1df4080-17ad-44ca-aeae-9d7de2419dbb', 'Module 4 user 1', 'module4-1@example.test', 1, '2026-11-04T00:00:00Z'),
    ('704137fe-93b3-4113-8e09-ed1df4d0e0ea', 'Module 4 user 2', 'module4-2@example.test', 1, '2026-11-04T00:00:00Z'),
    ('57172605-9d97-4245-9417-8714196b1d20', 'Module 4 user 3', 'module4-3@example.test', 1, '2026-11-04T00:00:00Z')
on conflict do nothing;

insert into public.subscriptions (user_id, plan, updated_at)
values
    ('e1df4080-17ad-44ca-aeae-9d7de2419dbb', 'BASE', '2026-11-04T00:00:00Z'),
    ('704137fe-93b3-4113-8e09-ed1df4d0e0ea', 'BASE', '2026-11-04T00:00:00Z'),
    ('57172605-9d97-4245-9417-8714196b1d20', 'BASE', '2026-11-04T00:00:00Z')
on conflict do nothing;

-- +goose Down
delete from public.subscriptions
where user_id in (
    'e1df4080-17ad-44ca-aeae-9d7de2419dbb',
    '704137fe-93b3-4113-8e09-ed1df4d0e0ea',
    '57172605-9d97-4245-9417-8714196b1d20'
);

delete from public.users
where id in (
    'e1df4080-17ad-44ca-aeae-9d7de2419dbb',
    '704137fe-93b3-4113-8e09-ed1df4d0e0ea',
    '57172605-9d97-4245-9417-8714196b1d20'
);
