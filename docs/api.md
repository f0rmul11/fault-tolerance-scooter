# HTTP/gRPC API

HTTP-шлюз: `http://localhost:18080`. Swagger: `/docs/`, схема: `/swagger.json`.
Внутренние сервисы предоставляют gRPC. На всех входах сохранены reflection и Echo.

| HTTP | Путь | RPC владельца |
|---|---|---|
| GET | `/v1/profiles/{user_id}` | `profile.profile.v1.ProfileService/GetProfile` |
| GET | `/v1/profiles` | `profile.profile.v1.ProfileService/ListProfiles` |
| PATCH | `/v1/profiles/{user_id}` | `profile.profile.v1.ProfileService/UpdateProfile` |
| GET | `/v1/scooters/{scooter_id}` | `scooter.scooter.v1.ScooterService/GetScooter` |
| GET | `/v1/scooters` | `scooter.scooter.v1.ScooterService/ListScooters` |
| PATCH | `/v1/scooters/{scooter_id}/availability` | `scooter.scooter.v1.ScooterService/UpdateScooterAvailability` |
| PATCH | `/v1/scooters/{scooter_id}/owner` | `scooter.scooter.v1.ScooterService/UpdateScooterOwner` |
| POST | `/v1/rentals` | `rental.rental.v1.RentalService/ReserveRental` |
| GET | `/v1/rentals/{rental_id}` | `rental.rental.v1.RentalService/GetRental` |
| POST | `/v1/rentals/{rental_id}:start` | `rental.rental.v1.RentalService/StartRental` |
| POST | `/v1/rentals/{rental_id}:finish` | `rental.rental.v1.RentalService/FinishRental` |
| POST | `/v1/rentals/{rental_id}:cancel` | `rental.rental.v1.RentalService/CancelRental` |
| GET | `/v1/rentals` | `rental.rental.v1.RentalService/ListRentals` |
| GET | `/v1/pricing/trips/{rental_id}` | `pricing.pricing.v1.PricingService/GetTrip` |
| GET | `/v1/pricing/tariffs` | `pricing.pricing.v1.PricingService/ListTariffs` |

Поля `user_id`, `rental_id`, `scooter_id`, `idempotency_key` — UUID. JSON `int64`
возвращается десятичными строками по protobuf JSON mapping. Profile использует enum
`SUBSCRIPTION_PLAN_BASE`, `SUBSCRIPTION_PLAN_PREMIUM`, `SUBSCRIPTION_PLAN_ULTRA`.
Для обновления профиля/самоката обязательна `expected_version`. Optional поля
не передавайте, если значение не меняется. У Scooter две отдельные ручки: availability
требует явный boolean is_active (включая false), owner — непустое название юрлица
до 200 символов. Обе требуют expected_version. Заряд и общий UpdateScooter удалены.

Request ID берётся из `x-request-id` либо создаётся шлюзом и возвращается в ответе.
Во внутренние изменяющие RPC Rental он передаётся полем `request_id`. Это не ключ
идемпотентности: после timeout повторяйте Reserve с тем же `idempotency_key`.

## Полная поездка вручную

Ни одна из этих операций не запускается автоматически из scooterctl.

```sh
USER_ID='17e1b7b2-0373-4356-afc3-257ccae69237'
SCOOTER_ID='3499a1fb-030c-47a8-bcd2-0ce8304d6a22'
KEY=$(uuidgen | tr '[:upper:]' '[:lower:]')
curl -s http://localhost:18080/v1/profiles
curl -s http://localhost:18080/v1/scooters
curl -s -X POST http://localhost:18080/v1/rentals \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER_ID\",\"scooter_id\":\"$SCOOTER_ID\",\"idempotency_key\":\"$KEY\"}"
```

Сохраните `rental.id` из ответа. Для последующих команд:

```sh
RENTAL_ID='<rental.id из предыдущего ответа>'
curl -s -X POST "http://localhost:18080/v1/rentals/$RENTAL_ID:start" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER_ID\"}"
curl -s "http://localhost:18080/v1/pricing/trips/$RENTAL_ID?user_id=$USER_ID"
curl -s -X POST "http://localhost:18080/v1/rentals/$RENTAL_ID:finish" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER_ID\"}"
curl -s "http://localhost:18080/v1/pricing/trips/$RENTAL_ID?user_id=$USER_ID"
```

Pricing обновляется асинхронно. До первого события возможен 404, затем RESERVED,
ACTIVE с авансом и COMPLETED с итогом. У BASE короткая поездка стоит 1000 копеек:
5000 копеек аванса, затем −4000 копеек корректировки, итоговое списание 1000.
`total_minor`/`adjustment_minor` отсутствуют до итогового расчёта.

## Ошибки и повторы

JSON ошибок — стандартный gRPC Status (`code`, `message`, `details`); сообщения
по-русски. HTTP 422 — неверные параметры, 404 — отсутствующая запись, 403 — чужая
аренда/цена, 409 — бизнес-отказ, CAS или конфликт ключа, 503/504 — недоступность/таймаут.
`x-error-code` приходит от владельца. При потере ответа нельзя считать операцию
отменённой: перечитайте аренду либо повторите ту же команду. Start/finish/cancel
идемпотентны для уже достигнутого состояния. После completed новый start запрещён.

Авторизация учебная: user_id задаётся явно, JWT/IAM нет. Прямой gRPC проходит
те же доменные проверки, что и HTTP. Не размещайте стенд как публичный сервис.
