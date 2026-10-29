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
