insert into public.tariffs (plan, rate_minor, advance_minor)
values
	('BASE', 1000, 5000),
	('PREMIUM', 700, 3000),
	('ULTRA', 400, 1000)
on conflict do nothing;
