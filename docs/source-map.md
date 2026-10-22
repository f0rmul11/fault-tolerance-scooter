# Происхождение реализации

Источник переноса: существующий HotelFlow. Это автономная копия технических
частей, а не Go-зависимость от соседнего репозитория. Исходный hotel не изменялся.

| Источник HotelFlow | ScooterFlow | Адаптация |
|---|---|---|
| services/api-gateway | services/api-gateway | Clay, probes, shutdown, rtconfig, requestid, metadata; новые API и прямые клиенты |
| services/hotels | profile, scooter | App/init/config, pgx/scany, transaction, xo templates, ошибки; компактный новый домен |
| services/booking | rental | event/pipe/outbox/worker/producer; новая аренда и один топик |
| services/check-in | pricing | inbox/worker/consumer/router; расчёт и события аренды |
| hotelctl | scooterctl | Cobra/factory, state/lock, Docker SDK, client-go, immutable migration Jobs, lifecycle, reset |
| deploy/booking-dependencies | deploy/dependencies | 4 PostgreSQL, Kafka, AKHQ, etcd; новый namespace и владельцы |

Общий функционал утилит: обнаружение проекта, конфигурация, сборка и импорт
образов, владение инфраструктурой, readiness, диагностика, управление репликами,
логи, публикация портов, миграции/seed, reset и runtime config. Скопирован внутрь
scooterctl с сохранением пакетов. Внешнего общего репозитория/Go-модуля нет,
поэтому каждый стенд можно клонировать и разворачивать независимо.
Бизнесовые smoke, hotel-сущности и Notification исключены.

Инструкции агентов (`AGENTS.md`, `.agents`, `.codex` и локальная копия
`docs/reference/hotel-AGENTS.md`) исключены из Git и остаются локально.
Общие доменные решения для репозитория описаны в спецификации ScooterFlow.
