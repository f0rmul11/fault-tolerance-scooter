# profile

Часть самостоятельного стенда ScooterFlow. Контракт: `api/profile`,
конфигурация: `config/config.yaml`, запуск: `go run ./cmd`, unit/race: `go test -race ./...`.

Архитектура и сценарии описаны в [общей спецификации](../../docs/specification.md).
Сборка/деплой через scooterctl из корня. Три probe endpoint на admin HTTP:
`/healthcheck/live`, `/healthcheck/ready`, `/healthcheck/startup`.
