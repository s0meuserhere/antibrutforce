# Анти-брутфорс

[![ci](https://github.com/s0meuserhere/antibrutforce/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/s0meuserhere/antibrutforce/actions/workflows/ci.yml)

Сервис вызывается перед авторизацией и решает, пропустить ли попытку входа. Он ограничивает частоту попыток по логину, паролю и IPv4 и ведёт whitelist и blacklist подсетей.

## Как принимается решение

- Адрес из whitelist пропускается, из blacklist отклоняется. Бакеты в обоих случаях не меняются. Если адрес есть в обоих списках, он пропускается.
- Остальные попытки учитываются в трёх бакетах: логина, пароля и IP. Попытка проходит, только если место нашлось в каждом из них.
- Бакет устроен как leaky bucket: ёмкость равна лимиту попыток в минуту, попытки вытекают равномерно, по `лимит / 60` в секунду. Лимиты по умолчанию: 10 на логин, 100 на пароль, 1000 на IP.
- Списки хранятся в PostgreSQL и кэшируются в Redis. Бакеты лежат в Redis и исчезают по TTL, когда опустеют. Пароль в Redis не попадает, в ключе бакета лежит его SHA-256.
- Проверка атомарна для всех экземпляров сервиса: бакеты читаются и пишутся в транзакции Redis `WATCH`/`MULTI`.

## Запуск

```bash
make run
```

Поднимает PostgreSQL на порту `5435`, Redis на `6380` и сервис на `50051`. Перед стартом сервиса контейнер `migrate` применяет миграции goose из `migrations`. `make down` останавливает контейнеры.

## Команды Makefile

| Команда | Что делает |
|---|---|
| `make build` | собирает `bin/antibf` и `bin/antibfcli` |
| `make test` | юнит-тесты: `go test -race -count 100 ./...` |
| `make integration-tests` | поднимает окружение из `deployments/docker-compose.test.yaml`, прогоняет тесты из `test/integration` и удаляет контейнеры |
| `make lint` | golangci-lint v2.15.0 по всему модулю, включая интеграционные тесты |
| `make generate` | генерирует код gRPC из `api/antibf.v1.proto` в `gen/antibf/v1` |
| `make run`, `make down` | поднимают и останавливают окружение из `deployments/docker-compose.yaml` |

Инструменты (golangci-lint, `protoc-gen-go`, `protoc-gen-go-grpc`) ставятся в `bin/tools` закреплённых версий. Для `make generate` нужен `protoc`: на macOS `brew install protobuf`, на Debian и Ubuntu `sudo apt install protobuf-compiler`, остальные способы описаны на https://protobuf.dev/installation/.

Новая миграция создаётся командой `goose -dir migrations create <имя> sql`.

## Настройки

Сервис и CLI читают настройки из переменных окружения. Файл с переменными задаётся флагом `-config` у `antibf` и `-c` или `--config` у `antibfcli`, по умолчанию `./configs/antibf.env`. Файл необязателен, а переменные окружения важнее значений из него.

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `LIMIT_LOGIN` | `10` | попыток в минуту на логин |
| `LIMIT_PASSWORD` | `100` | попыток в минуту на пароль |
| `LIMIT_IP` | `1000` | попыток в минуту на IP |
| `REQUEST_TIMEOUT` | `1s` | срок одного запроса к сервису |
| `NETLIST_CACHE_TTL` | `1m` | насколько кэш списков может отстать от PostgreSQL |
| `DB_HOST`, `DB_PORT` | `localhost`, `5432` | адрес PostgreSQL |
| `DB_USER`, `DB_PASSWORD`, `DB_NAME` | - | обязательны |
| `DB_SSL` | `disable` | `sslmode` подключения |
| `REDIS_HOST`, `REDIS_PORT` | `localhost`, `6379` | адрес Redis |
| `REDIS_PASSWORD`, `REDIS_DB` | пусто, `0` | доступ к Redis |
| `GRPC_HOST`, `GRPC_PORT` | `0.0.0.0`, `50051` | адрес сервера, только для `antibf` |
| `ENVIRONMENT` | `local` | `local` - логи текстом, иначе JSON, только для `antibf` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` или `error`, только для `antibf` |

## API

gRPC, пакет `antibf.v1`, сервис `AntiBruteforceService`. Контракт лежит в `api/antibf.v1.proto`.

| Метод | Запрос | Ответ |
|---|---|---|
| `CheckV1` | `login`, `password`, `ip` | `ok` - пропустить ли попытку |
| `ResetV1` | `login` и/или `ip` | сбрасывает бакеты логина и IP, бакет пароля остаётся |
| `AddToBlacklistV1`, `RemoveFromBlacklistV1` | `subnet` | добавляет или удаляет подсеть в blacklist |
| `AddToWhitelistV1`, `RemoveFromWhitelistV1` | `subnet` | добавляет или удаляет подсеть в whitelist |

Подсеть задаётся как IPv4 с маской, например `192.1.1.0/25`. Адрес хоста обнуляется: `192.1.1.10/25` сохраняется как `192.1.1.0/25`. Повторное добавление и удаление отсутствующей подсети ошибкой не считаются.

Ошибки:

- Неверный запрос получает код `InvalidArgument`. В деталях статуса лежит `google.rpc.ErrorInfo`, его `reason` - имя значения из enum `ErrReason` ответа метода, например `EMPTY_LOGIN`, `INVALID_IP` или `INVALID_SUBNET`. Список причин у каждого ответа свой, он описан в proto.
- Сбой хранилища получает код `Internal` без подробностей, сама ошибка пишется в лог сервиса.
- В ответе с кодом `OK` причины нет: в самих ответах поля `reason` нет.

## CLI

`antibfcli` собирает тот же сервис, что и сервер, и вызывает его напрямую, без gRPC. Ему нужны те же PostgreSQL и Redis.

```bash
./bin/antibfcli check --login admin --password secret --ip 192.0.2.1
./bin/antibfcli reset --login admin --ip 192.0.2.1
./bin/antibfcli blacklist add 192.1.1.0/25
./bin/antibfcli blacklist remove 192.1.1.0/25
./bin/antibfcli whitelist add 10.0.0.0/8
./bin/antibfcli whitelist remove 10.0.0.0/8
```

Файл настроек задаётся флагом `-c` или `--config`. Файл по умолчанию подходит к окружению `make run`.

## Структура

| Каталог | Что внутри |
|---|---|
| `api` | контракт gRPC |
| `gen/antibf/v1` | сгенерированный код gRPC |
| `cmd/antibf`, `cmd/antibfcli` | сервер и CLI |
| `internal/domain` | бизнес-логика: бакет, попытка, лимиты, подсети и списки |
| `internal/service/guard` | сценарии сервиса |
| `internal/repository` | хранение бакетов в Redis и списков в PostgreSQL с кэшем в Redis |
| `internal/grpcserver` | хендлер gRPC: запросы в вызовы сервиса, ошибки в код и `reason` |
| `internal/app`, `internal/config`, `internal/logger`, `internal/infrastructure` | сборка сервиса, настройки, логи и подключения |
| `migrations` | миграции goose |
| `test/integration` | интеграционные тесты под тегом `integration` |
