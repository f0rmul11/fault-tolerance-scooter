# Правила разработки ScooterFlow

Перед изменениями прочитай `docs/specification.md` и `docs/implementation-status.md`.
Структура и технологии перенесены из HotelFlow. Обязательны навыки
`.agents/skills/hotelflow-domain-usecase`, `hotelflow-storage`,
`hotelflow-contracts-events`, `hotelflow-errors-tests` и их тематические references.
Исходные правила сохранены в `docs/reference/hotel-AGENTS.md`.

Общие требования исходника действуют: Clean Architecture/DDD, узкие порты,
ручной DI в init/registry, отдельный use case и файл обработчика на сценарий,
DAO с embedded generated xo, With/Option, CAS, transaction.Exec,
event.Buffer/Flusher и скопированные outbox/inbox/worker/connectors.
Ошибки копируют error.go/option.go исходника. Go многострочный, gofmt с табами;
proto форматируется Buf (два пробела) с русскими комментариями.

Предметную область HotelFlow, роли гостя/ресепшн, GiST периодов номеров,
шесть сервисов, два топика, hotelctl и абсолютные пути исходника не переносить
в новый домен. Их заменяет `docs/specification.md`. Все пять сервисов и
scooterctl автономны; ссылок на исходный репозиторий во время сборки нет.
Только четыре владельца данных имеют свои PostgreSQL; gateway без БД.
Один прикладной топик rental.events.v1. Метрик, Redis, RPC retry нет.
В scooterctl нет smoke и бизнес-сценариев. up только строит/деплоит,
применяет миграции/seed и ждёт Kubernetes readiness.

Генерация proto только Buf, generated-файлы вручную не редактируются.
XO генерируется закреплённым инструментом из мигрированной изолированной БД.
Тесты: go vet и go test -race каждого модуля; проверять реальную генерацию,
SQL и запуск отдельно, не выдавать unit-тесты за интеграционные проверки.
