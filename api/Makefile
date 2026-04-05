APP     := server
CMD     := ./cmd/server
BIN     := ./bin/$(APP)
DB_URL  ?= $(shell grep DATABASE_URL .env | cut -d '=' -f2-)

.PHONY: run build dev sqlc migrate migrate-down migrate-status migrate-create lint vet clean

## Run server (with live .env)
run:
	go run $(CMD)

## Build binary to ./bin/server
build:
	go build -o $(BIN) $(CMD)

## Generate sqlc code
sqlc:
	sqlc generate

## Run DB migrations up
migrate:
	goose -dir internal/database/migrations postgres "$(DB_URL)" up

## Roll back last migration
migrate-down:
	goose -dir internal/database/migrations postgres "$(DB_URL)" down

## Show migration status
migrate-status:
	goose -dir internal/database/migrations postgres "$(DB_URL)" status

## Create new migration: make migrate-create name=add_sessions
migrate-create:
	goose -dir internal/database/migrations create $(name) sql

## Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

## Vet
vet:
	go vet ./...

## Remove build artifacts
clean:
	rm -rf ./bin
