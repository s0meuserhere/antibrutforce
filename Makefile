BIN_DIR := ./bin
TOOLS_DIR := $(abspath $(BIN_DIR)/tools)

GOLANGCI_VERSION := v2.15.0
PROTOC_GEN_GO_VERSION := v1.36.12
PROTOC_GEN_GO_GRPC_VERSION := v1.6.2

.PHONY: build test lint install-lint-deps install-proto-deps generate run down

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/antibf ./cmd/antibf
	go build -o $(BIN_DIR)/antibfcli ./cmd/antibfcli

test:
	go test -race -count 100 ./...

install-lint-deps:
	@$(TOOLS_DIR)/golangci-lint version 2>/dev/null | grep -q "version $(GOLANGCI_VERSION:v%=%) " \
		|| GOBIN=$(TOOLS_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

lint: install-lint-deps
	$(TOOLS_DIR)/golangci-lint run ./...

install-proto-deps:
	@command -v protoc > /dev/null || (echo "protoc not found: brew install protobuf (macOS), sudo apt install protobuf-compiler (Debian, Ubuntu), https://protobuf.dev/installation/" && exit 1)
	@$(TOOLS_DIR)/protoc-gen-go --version 2>/dev/null | grep -q " $(PROTOC_GEN_GO_VERSION)$$" \
		|| GOBIN=$(TOOLS_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@$(TOOLS_DIR)/protoc-gen-go-grpc --version 2>/dev/null | grep -q " $(PROTOC_GEN_GO_GRPC_VERSION:v%=%)$$" \
		|| GOBIN=$(TOOLS_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

generate: install-proto-deps
	mkdir -p gen/antibf/v1
	PATH="$(TOOLS_DIR):$$PATH" protoc -I api \
		--go_out=gen/antibf/v1 --go_opt=paths=source_relative \
		--go-grpc_out=gen/antibf/v1 --go-grpc_opt=paths=source_relative \
		api/antibf.v1.proto

run:
	docker compose -f deployments/docker-compose.yaml up --build

down:
	docker compose -f deployments/docker-compose.yaml down
