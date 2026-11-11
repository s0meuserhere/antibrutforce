FROM golang:1.25.7-alpine AS build

# Собираем goose только с драйвером PostgreSQL.
RUN CGO_ENABLED=0 go install -trimpath -ldflags="-s -w" \
    -tags='no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb' \
    github.com/pressly/goose/v3/cmd/goose@v3.27.3

FROM scratch

COPY --from=build /go/bin/goose /goose
COPY migrations /migrations

ENV GOOSE_DRIVER=postgres \
    GOOSE_MIGRATION_DIR=/migrations

USER 65534:65534

ENTRYPOINT ["/goose"]
CMD ["up"]
