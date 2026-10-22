insert into public.scooters (id, number, brand, is_active, owner, ridden_seconds, version, updated_at)
values
	('3499a1fb-030c-47a8-bcd2-0ce8304d6a22', 'SC-0001', 'Ninebot', true, 'ООО «Учебный прокат»', 7200, 1, '2026-10-22T00:00:00Z'),
	('85ff7871-fe56-4eaf-ae15-a4187a99177a', 'SC-0002', 'Xiaomi', true, 'ООО «Учебный прокат»', 14400, 1, '2026-10-22T00:00:00Z'),
	('d755f5a3-a700-48e3-8766-3bbc8a4f47d3', 'SC-0003', 'Ninebot', false, 'ООО «Учебный прокат»', 3600, 1, '2026-10-22T00:00:00Z'),
	('53226fdd-3323-4b2d-9bad-4f782f2b0885', 'SC-0004', 'Kugoo', false, 'ООО «Учебный прокат»', 1800, 1, '2026-10-22T00:00:00Z')
on conflict do nothing;
