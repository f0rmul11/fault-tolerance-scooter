# pricing

Часть самостоятельного стенда ScooterFlow. Контракт: `api/pricing`,
конфигурация: `config/config.yaml`, запуск: `go run ./cmd`, unit/race: `go test -race ./...`.

Архитектура и сценарии описаны в [общей спецификации](../../docs/specification.md).
Сборка/деплой через scooterctl из корня. Три probe endpoint на admin HTTP:
`/healthcheck/live`, `/healthcheck/ready`, `/healthcheck/startup`.

Для модуля 2 каталог тарифов читается из `tariffs.json` в корне сервиса.
`App.New` запускает фоновую загрузку с имитацией прогрева 30 секунд;
`TARIFFS_FILE` задаёт путь к файлу (`tariffs.json`). Поездки и inbox остаются в PostgreSQL.
В исходном задании readiness намеренно не учитывает результат загрузки;
административные cordon/uncordon добавляет ученик. Pricing запускается в трёх репликах.
