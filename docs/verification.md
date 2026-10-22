# Проверка ScooterFlow

Дата: 22 октября 2026. Проверки выполнялись на macOS arm64, Go 1.26.1,
Docker/OrbStack, отдельном k3s `scooterflow-k3s` v1.35.2-k3s1.

| Проверка | Результат |
|---|---|
| Buf format/lint/build/generate | Все владельцы и gateway, включая внешних клиентов; закреплённый Docker toolchain |
| PostgreSQL-миграции и xo | Четыре отдельные codegen БД; фактическая генерация not-for-prod/xo v0.0.1 |
| `make test vet` | Пройдено отдельно в каждом из шести Go-модулей; race включён |
| Lifecycle | Сохранённые HTTP/gRPC/reflection/probes/shutdown-тесты Profile и Scooter; runtime tests шлюза |
| `scooterctl up` | Развёрнуты пять приложений, четыре PostgreSQL, Kafka, AKHQ, etcd; readiness достигнута |
| Полный HTTP цикл | RESERVED → ACTIVE → COMPLETED; цена/аванс/возврат проверены через настоящий Kafka/inbox |
| Конкуренция | Два пользователя/один самокат и один пользователь/два самоката: ровно один победитель |
| Идемпотентность | Replay reserve/start/finish без нового состояния; другой состав ключа получает 409 |
| Profile CAS | Подписка обновляется; устаревшая версия не перезаписывает профиль |
| Optional значения | Явный charge_percent=0 проходит через шлюз; отсутствие поля сохраняет значение |
| Тарифный снимок | Изменение подписки после reserved не изменяет цену текущей поездки |
| Pricing recovery | При остановке Pricing аренда проходит все шаги; после запуска итог рассчитан |
| Kafka redelivery | Три outbox-события доставлены повторно с прежними correlation ID: inbox остался из 3 строк, цена не изменилась |
| Реальный rollback | Ошибка после вставки аренды и outbox откатила обе записи в PostgreSQL |
| Реальный CAS | Обновление устаревшей версии аренды отклонено |

Unit-тесты отдельно проверяют границы заряда, недопустимые переходы и время,
неизменность состояния при отказе, округление начатых минут, переполнение денег,
пропуск предшествующего события, повтор аванса и отсутствие fallback тарифа.
Они не подменяют перечисленные проверки настоящих PostgreSQL/Kafka/k3s.

Опциональная воспроизводимая проверка транзакции (БД должна быть изолированной,
с уже применёнными миграциями Rental):

```sh
export SCOOTER_RENTAL_TEST_DSN='postgres://user:password@127.0.0.1:5432/rental_test?sslmode=disable'
go test -race -tags integration ./services/rental/internal/application/usecase/rental/reserve -run TestPostgresAtomicOutboxAndCAS -v
```

Локальные логи прогонов хранятся в `.local/verification` и не входят в Git.
Проверки не претендуют на нагрузочный benchmark, HA или exactly-once.

## Проверка перед передачей

Все образы заново собраны и развёрнуты из итогового каталога
`/Users/d.bulykin/Desktop/faut-tolerance-course/scooter`.
Повторный `make test vet` завершился с кодом 0 во всех шести модулях;
полный HTTP прогон подтвердил 9 групп проверок, включая восстановление Pricing.

При остановленном Pricing шлюз может вернуть 503 либо ждать до deadline клиента:
собственный RPC deadline в исходном шлюзе отсутствует и в этом стенде не добавлен.
Проверка отказа использует ограничение времени на стороне клиента и гарантированно
возвращает Pricing в работу перед ожиданием итогового расчёта.

`reset --yes` завершился с кодом 0. Прямой SQL через встроенный kubectl подтвердил:
users=3, subscriptions=3, scooters=4, tariffs=3;
rentals=0, trips=0, outbox=0, inbox=0.
Kafka topics: rental.events.v1 и служебный __consumer_offsets.
`status` показал 1/1 для всех пяти приложений и семи инфраструктурных компонентов,
`doctor` подтвердил Docker и готовый узел k3s. HTTP-списки вернули seed,
Swagger `/docs/` ответил 200. Состояние оставлено запущенным и готовым к занятиям.

Итоговые локальные файлы: `.local/verification/final-http.log`,
`delivered-checks.log`, `final-reset.log`, `final-status.json`, `delivery.json`.
