BIN_DIR := ./bin
GOLANGCI_VERSION := v2.15.0
GOLANGCI_BIN := $(shell go env GOPATH)/bin/golangci-lint

.PHONY: build test lint install-lint-deps run down

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/antibf ./cmd/antibf
	go build -o $(BIN_DIR)/antibfcli ./cmd/antibfcli

test:
	go test -race -count 100 ./...

install-lint-deps:
	@$(GOLANGCI_BIN) version 2>/dev/null | grep -q "version $(GOLANGCI_VERSION:v%=%) " \
		|| go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

lint: install-lint-deps
	$(GOLANGCI_BIN) run ./...

run:
	docker compose -f deployments/docker-compose.yaml up --build

down:
	docker compose -f deployments/docker-compose.yaml down
