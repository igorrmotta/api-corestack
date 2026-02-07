.PHONY: build run-server run-worker test lint sqlc-generate

build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

run-server:
	go run ./cmd/server

run-worker:
	go run ./cmd/worker

test:
	go test ./...

test-unit:
	go test ./internal/service/... ./internal/handler/...

test-integration:
	go test ./internal/repository/...

lint:
	golangci-lint run ./...

sqlc-generate:
	sqlc generate
