# ScooterFlow — стенд домашних заданий

Самостоятельный стенд аренды самокатов на Go, gRPC, Clay, PostgreSQL, Kafka
и k3s. Пять сервисов, четыре независимые БД и один прикладной Kafka-топик.
Никакие файлы, сервисы или данные HotelFlow для запуска не нужны.

## Запуск

Требуются Go 1.26.1 и запущенный Docker с Linux-контейнерами (Docker Desktop
или OrbStack). Docker должен иметь достаточно памяти для k3s, четырёх PostgreSQL,
Kafka и сборки Go. Интернет нужен для первого получения образов и Go-зависимостей.
Локальные kubectl, k3s, Kafka и PostgreSQL устанавливать не требуется.

```sh
make cli
./bin/scooterctl doctor
./bin/scooterctl up --timeout 15m
./bin/scooterctl status
```

`up` создаёт собственный `scooterflow-k3s`, применяет миграции и идемпотентный
статический seed, разворачивает сервисы и ждёт readiness Kubernetes.
Бизнес-запросов и автоматических smoke в этой команде нет.

| Компонент | Локальный адрес |
|---|---|
| HTTP API / Swagger | http://localhost:18080/docs/ |
| Gateway gRPC | 127.0.0.1:19090 |
| Gateway admin | http://localhost:18081/healthcheck/ready |
| Profile / Scooter / Rental / Pricing gRPC | 19091 / 19092 / 19093 / 19094 |
| Profile / Scooter / Rental / Pricing admin | 18082 / 18083 / 18084 / 18085 |
| Kafka UI | http://localhost:18090 |
| Rental / Pricing / Profile / Scooter PostgreSQL | 25440 / 25441 / 25442 / 25443 |
| Runtime config etcd | 127.0.0.1:12379 |

Адреса БД и синтетические учебные учётные данные: `scooterctl access`.
Сервисы внутри k3s используют собственные DNS-имена и обычные порты 9090/5432.
Kafka доступна только внутри кластера по `kafka:9092`.

## Работа со стендом

```sh
./bin/scooterctl build rental pricing --timeout 10m
./bin/scooterctl deploy rental pricing --timeout 10m
./bin/scooterctl logs rental
./bin/scooterctl stop pricing
./bin/scooterctl start pricing
./bin/scooterctl restart rental
./bin/scooterctl components
./bin/scooterctl kubectl -n scooterflow get pods
./bin/scooterctl stop
./bin/scooterctl up --timeout 15m
```

`stop` сохраняет данные. `reset --yes` согласованно очищает четыре БД,
восстанавливает seed, пересоздаёт единственный прикладной топик и сбрасывает
consumer group; возобновляет прежнее количество реплик. Прерванный reset
продолжается повторением команды. `destroy --yes` удаляет кластер и его данные.
Учебные бизнес-сценарии и smoke-команды в scooterctl отсутствуют.

## Предметная область и API

[Спецификация](docs/specification.md), [API и примеры](docs/api.md),
[учебные UUID](docs/seed.md), [события](api/asyncapi.yaml),
[происхождение общего кода](docs/source-map.md),
[проверки реализации](docs/verification.md).

Rental управляет жизненным циклом RESERVED → ACTIVE → COMPLETED и отменой
до старта. Scooter остаётся справочником; занятость хранится только в Rental.
Pricing получает события асинхронно, фиксирует тариф через Profile, учитывает
аванс и рассчитывает полную стоимость по начатым минутам. Деньги — копейки RUB.
Это учебный финансовый учёт без списаний с реальных банковских карт.

Структура владельца: `internal/domain/entity`, `application/usecase`,
`infrastructure/storage/<entity>/dao`, `infrastructure/gateway`, `app`,
`events` (у издателя), `pkg`; ручная сборка в `internal/init.go` и registry.
Шлюз содержит только транспорт и generated gRPC-клиенты.

## Проверки и генерация

```sh
make test
make vet
make generate
```

Generated protobuf и xo закоммичены вместе с источниками: обычной сборке
генераторы не нужны. Все proto локальны сервису; внешние клиенты читаются
из API владельца через Buf. Версии инструментов закреплены в Dockerfile.codegen.
`make generate-local` требует тех же плагинов на PATH.

При недоступном Buf Registry разрешён внешний кеш импортов:

```sh
services/api-gateway/tools/prepare-proto-imports.sh /tmp/scooter-proto-imports
make -C services/api-gateway generate-offline PROTO_IMPORTS=/tmp/scooter-proto-imports
```

XO генерируется только из отдельно мигрированной БД своего сервиса:

```sh
make xo-tools
export POSTGRES_DSN='postgres://user:password@localhost:5432/codegen?sslmode=disable'
make migrate SERVICE=profile
make generate-xo SERVICE=profile
```

Генерация не сбрасывает БД. Для runtime-конфига шлюза после изменения
`config/rtconfig.yaml`: `cd services/api-gateway/internal/config && go run ./generate`.
`ScootersListEnabled` сохранён как учебный пример reader/writer, API им пока не ограничивается.

В репозитории нет Redis, автоматических RPC retry, circuit breaker, метрик,
HA-инфраструктуры или production IAM. HTTP принимает явные user_id в доверенной
локальной среде; проверка принадлежности аренды/расчёта выполняется владельцем.
