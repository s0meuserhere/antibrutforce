# Анти-брутфорс

Сервис ограничивает частоту попыток авторизации по логину, паролю и IPv4.

## Проверки

```bash
make build
make test
make lint
make run
make down
```

`make test` запускает `go test -race -count 100 ./...`. `make lint` ставит golangci-lint v2.15.0 и проверяет весь модуль. `make build` собирает `bin/antibf` и `bin/antibfcli` на Go 1.25.6. `make run` собирает образ сервиса и поднимает его на порту `50051` вместе с PostgreSQL на `5435` и Redis на `6380`. Перед стартом сервиса контейнер `migrate` применяет миграции goose из `migrations`. Новая миграция создаётся командой `goose -dir migrations create <имя> sql`. `make down` останавливает контейнеры.
