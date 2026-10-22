# ScooterFlow: спецификация стенда домашних заданий

22 октября 2026. Самостоятельный репозиторий, пять Go-модулей сервисов и
Go CLI scooterctl. Повторяет архитектуру HotelFlow: Clay, gRPC, PostgreSQL,
Kafka, k3s в Docker, Buf, Goose, xo, ручной DI и Clean Architecture.

## Владение данными

| Сервис | Данные | Внешние зависимости |
|---|---|---|
| api-gateway | HTTP/gRPC proxy, request_id, transport metadata | Все четыре владельца |
| profile | users и subscriptions, имя, email, BASE/PREMIUM/ULTRA | Свой PostgreSQL |
| scooter | UUID, номер, марка, заряд 0–100%, наезд в секундах | Свой PostgreSQL |
| rental | Бронь, поездка, занятость пользователя/самоката, HTTP-ключ, outbox | Scooter gRPC, свой PostgreSQL, Kafka |
| pricing | Тарифы в копейках, снимок тарифа поездки, аванс и итог, inbox | Profile gRPC, свой PostgreSQL, Kafka |

Gateway не имеет БД, domain/usecase/storage. Четыре PostgreSQL физически
разделены; никакой SQL к чужой БД. Никаких зависимостей от hotel.
`PREMUIM` в постановке нормализовано до `PREMIUM` в контрактах.
Redis пока отсутствует. Нет реального эквайринга, кошелька и телеметрии:
pricing учитывает учебное списание/возврат в собственной БД, scooter хранит
переданное оператором состояние. Наезд отдаётся в секундах, часы = seconds/3600.

## Сценарии и инварианты

Пользователи и самокаты задаются seed; API читает список/карточку и обновляет
данные. Минимальные поля пользователя: UUID, имя, email; подписка отдельной
строкой users.id = subscriptions.user_id. Обновление профиля и подписки атомарно,
с optimistic version. Цены принадлежат только Pricing.

Rental: RESERVED → ACTIVE → COMPLETED; RESERVED → CANCELED.
Один пользователь и один самокат могут иметь только одну незавершённую аренду.
Это обеспечивают частичные уникальные индексы по статусам RESERVED/ACTIVE.
Справочник Scooter не хранит занятость. Наличие самоката в каталоге не означает
его доступность для бронирования. Бронирование проверяет заряд ≥20%,
сохраняет снимок номера и марки. Повторный запрос того же пользователя с тем же
UUID idempotency_key и scooter_id возвращает сохранённую аренду до обращения
к Scooter. Другой scooter_id с тем же ключом — Conflict. При конкурентной вставке
проигравший получает Conflict и может повторить тот же ключ.

Start/finish/cancel принимают rental_id и user_id, проверяют владение,
выполняют CAS и сохраняют событие с изменением состояния одной транзакцией.
Повтор уже выполненного перехода возвращает текущее состояние без новой версии
и события. Start после completed/canceled запрещён. Finish до start запрещён.
Cancel активной поездки запрещён: нужна операция finish. Зарезервированная
поездка не тарифицируется. Автоистечения брони нет: отмена явная.
Смена подписки или полей каталога не меняет сохранённые снимки аренды/стоимости.
UUID v4 создаются через uuid.NewV4, время UTC с точностью микросекунд.

## События и стоимость

Один прикладной топик `rental.events.v1`, key = UUID rental_id, hash partitioner.
События `rental.reserved.v1`, `rental.started.v1`, `rental.completed.v1`,
`rental.canceled.v1`. Envelope: correlation_id (UUID события), event_type,
rental_id, rental_version, occurred_at, request_id, payload. Payload содержит
user_id, scooter_id, номер/марку, статус и фактические времена аренды.
Correlation ID создаётся один раз до сохранения outbox, совпадает с
x-correlation-id; x-event-type совпадает с envelope. Повторы relay сохраняют ID.
Request ID — только трассировка, HTTP idempotency_key — отдельный ключ команды.

Rental сохраняет данные и outbox атомарно. Relay копируется из исходного стенда;
его транзакция может охватывать отправку Kafka. Pricing сохраняет сообщение в
inbox до подтверждения offset. Inbox дедуплицирует (correlation_id, topic).
Pricing имеет отдельный use case для каждого типа события. Порядок Kafka сам
по себе недостаточен: workers inbox/outbox могут переупорядочить обработку.
Pricing применяет последовательные rental_version. Пропущенная предшествующая
версия даёт технический retry через inbox, старые версии не откатывают итог.
CAS и уникальность rental_id защищают от параллельной обработки.

Первое событие резервирования: Pricing читает **текущую подписку на момент
обработки события** у Profile и сохраняет снимок тарифа. История подписок и
тарифов не ведётся; это явно выбранная семантика асинхронного учебного стенда.
Недоступность Profile оставляет событие в inbox для повторной обработки.
Отсутствующий пользователь не подменяется BASE и требует исправления данных.
Rental не вызывает Profile, поэтому принятие брони не доказывает готовность цены.
Чтение цены до обработки reserved возвращает NotFound; клиент повторяет чтение.

Seed цен: BASE 1000 коп/мин, аванс 5000 коп; PREMIUM 700 и 3000; ULTRA 400 и 1000.
Валюта RUB. Started фиксирует аванс как учебное списание; повтор не списывает
его заново. Completed: минуты = ceil((finished_at-started_at)/60 секунд),
минимум 1; total = minutes × rate. Итоговое изменение = total − advance:
положительное означает доплату, отрицательное — возврат. Net charged = total.
Если finish не получен, charged остаётся advance; итог ещё неизвестен.
Canceled из RESERVED имеет нулевую стоимость и не списывает аванс.
Деньги только int64 копейки, без float. Расчёт проверяет переполнение.

## Контракты и ошибки

Источники proto: services/<owner>/api/<owner>/<area>/v1. Gateway владеет
отдельными публичными proto с HTTP annotations. Внешние клиенты генерируются
из владельца в internal/pkg/pb/external. Список стабильно сортируется по ID,
limit default 20/max 100, offset ≥0. Все ID в wire — UUID строки.
Optional поля обновления различают отсутствие и нулевое значение.
Обновления profile/scooter требуют expected_version; конфликт не перетирает
чужое обновление. Статусы и тарифы перечислены контрактом.

InvalidArgument → HTTP 422, NotFound →404, FailedPrecondition/Aborted →409,
Unavailable →503, DeadlineExceeded →504. Русские сообщения владельца и
x-error-code проходят через общий interceptor и Clay. Gateway не содержит
бизнес-валидации/таймаутов/retry/fallback. Транспорт demo-идентичности не является
production IAM: user_id передаётся явно, стенд работает в доверенной локальной
среде. Предположение не заменяет проверку владения арендой по user_id.

## Развёртывание и CLI

Свои namespace scooterflow, k3s container/volume/config, порты вне диапазонов
HotelFlow. У каждого владельца свой Deployment/Service/PVC/Secret PostgreSQL.
Kafka одна (KRaft), AKHQ для наблюдения. Скопированный runtime-config gateway
использует отдельный config-etcd. Все таблицы в public. Миграции выполняет Job
до запуска приложений; seed идемпотентен и не сбрасывает пользовательские данные.

scooterctl использует дерево Cobra, factory, project store/lock, Docker SDK и
Kubernetes client-go исходного hotelctl. Общие возможности: up/build/deploy,
start/stop/restart/components/status/logs/doctor/access/hosts/kubectl/destroy,
инфраструктурный reset и runtime config. В нём нет smoke и бизнес-сценариев.
Копии общего инфраструктурного кода независимы: клонирование одного репозитория
достаточно. `up` не выполняет HTTP/Echo/бизнес-проверок.
Один узел и по одной БД/брокеру предназначены для курса и не обеспечивают HA.
