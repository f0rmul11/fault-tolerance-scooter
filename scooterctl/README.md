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
