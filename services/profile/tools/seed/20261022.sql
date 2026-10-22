begin;
insert into public.users (id, name, email, version, updated_at)
values
	('17e1b7b2-0373-4356-afc3-257ccae69237', 'Ученик 1', 'student1@example.test', 1, '2026-10-22T00:00:00Z'),
	('e58a0390-e434-4b99-9a19-8c0e09ba0b00', 'Ученик 2', 'student2@example.test', 1, '2026-10-22T00:00:00Z'),
	('205bfad3-44aa-48a6-b395-3900d36f273d', 'Ученик 3', 'student3@example.test', 1, '2026-10-22T00:00:00Z')
on conflict do nothing;

insert into public.subscriptions (user_id, plan, updated_at)
values
	('17e1b7b2-0373-4356-afc3-257ccae69237', 'BASE', '2026-10-22T00:00:00Z'),
	('e58a0390-e434-4b99-9a19-8c0e09ba0b00', 'PREMIUM', '2026-10-22T00:00:00Z'),
	('205bfad3-44aa-48a6-b395-3900d36f273d', 'ULTRA', '2026-10-22T00:00:00Z')
on conflict do nothing;
commit;
