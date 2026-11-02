# scooterctl

Инфраструктурный CLI самостоятельного ScooterFlow, перенесённый из hotelctl.
Дерево Cobra → factory → stand.Manager → Docker SDK / client-go. Project store
хранит состояние и lock в `.local/scooterctl`; owner labels защищают чужие
контейнеры, сеть и том. `scooter.yaml` определяется поиском вверх от текущей папки.

Сборка из корня: `make cli`. Справка: `./bin/scooterctl --help`.

| Команда | Назначение |
|---|---|
| up | k3s, сборка, миграции/seed, деплой, ожидание readiness |
| build / deploy | Пересборка / применение выбранных сервисов |
| start / stop / restart | Управление сервисами и зависимостями |
| status / components / logs | Наблюдение за ресурсами |
| fault | Ошибки и задержки на выбранном RPC; Ctrl+C снимает сбой |
| load | gRPC-нагрузка через Framer и отчёты |
| doctor | Конфигурация, Docker, порты, готовность узла |
| access / hosts | Локальные адреса и опциональные записи hosts |
| kubectl | kubectl из контейнера k3s с конфигом этого проекта |
| config get / config set | Чтение/обновление runtime config в отдельном etcd |
| reset --yes | Очистка учебных БД и топика, восстановление seed и реплик |
| destroy --yes | Удаление принадлежащих стенду ресурсов и данных |

Smoke и бизнес-команд нет. Ни up, ни reset не создают аренд и не вызывают API.
По умолчанию порты отличаются от HotelFlow; общий сервер Docker допустим,
но Kubernetes, БД, Kafka, credentials и runtime config независимы.

```sh
./bin/scooterctl config get api-gateway scooters_list_enabled
./bin/scooterctl config set api-gateway scooters_list_enabled true
```

Этот флаг — пример чтения динамической конфигурации; gate API пока не подключён.
Полное описание запуска и портов находится в [README стенда](../README.md).

## Домашняя работа: модуль 2

```sh
./bin/scooterctl status pricing --watch
./bin/scooterctl homework module-2 run readiness
./bin/scooterctl homework module-2 run shutdown
./bin/scooterctl homework module-2 run cordon
```

`status pricing` показывает отдельные экземпляры; `--watch` обновляет таблицу раз в секунду.
`readiness` выполняет rollout под HTTP-запросами через Gateway и проверяет ошибку файла на изолированном pod.
`shutdown` завершает один Pricing и выводит readiness/liveness и журнал остановки.
`cordon` проверяет ручки, маршрутизацию новых gRPC-соединений, прогрев и ошибку файла.
Для маршрутизации используются новые соединения прямо к Service Pricing; UID и рестарты проверяются отдельно.

Все сценарии ждут исходного прогрева; временные pod удаляются автоматически. Для проверки нужны хотя бы две реплики, в манифесте Pricing — три.
Результаты выводятся в терминал: счётчики полных/пустых ответов и ошибок, распределение новых соединений по pod, проверки UID и рестартов. Файлов отчётов нет.
При нарушении критерия команда возвращает ненулевой код. После исправлений выполните `build pricing`, `deploy pricing` и повторите ту же проверку.

## Сбои и нагрузка: модуль 3

```sh
# Терминал 1: 5% попыток GetScooter получают Unavailable до обработки.
./bin/scooterctl fault api-gateway scooter --method GetScooter --unavailable --percent 5 --duration 15m

# Терминал 2: пользовательский поток; генератор сам не делает retry.
./bin/scooterctl load --addr 127.0.0.1:19090 \
  --method api_gateway.scooter.v1.ScooterService/GetScooter \
  --data-file homeworks/module-3/get-scooter.json \
  --rps 20 --duration 10m --request-timeout 3s

# Терминал 3: каждая завершённая исходящая попытка, включая искусственный отказ.
./bin/scooterctl logs api-gateway --outgoing -f --tail 0
```

`fault` поддерживает связь `api-gateway → scooter` и метод `GetScooter`.
Для задержки: `--slowdown 500ms` вместо `--unavailable`. Вероятность задаёт
`--percent` (по умолчанию 100). Команда ждёт окончания `--duration` и снимает
правило; Ctrl+C снимает его досрочно. После SIGKILL/обрыва используйте
`fault api-gateway scooter --clear`. Не перезапускайте стенд во время измерения.

`load` автоматически собирает закреплённый Framer (Go, Git и сеть нужны при первом
запуске), готовит запрос и сохраняет `report.phout`, `requests.tsv`, `summary.txt`
и `summary.json` в `.local/scooterctl/load`. `--report-dir` меняет родительскую
папку результатов. При прерывании отчёт содержит только завершённые запросы.
Служебные запросы reflection не входят в исходящие логи GetScooter.
Ретраи и бюджет в Gateway намеренно не реализованы: это задача ученика.

Полный сценарий — [домашняя работа модуля 3](../homeworks/module-3/README.md).
