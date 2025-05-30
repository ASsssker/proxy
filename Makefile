ifeq ("$(wildcard .env)","")
$(shell cp .env-example .env)
endif

include .env

## help: вывод информации о командах
.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


## up: псевдоним для "docker compose up -d --build"
.PHONY:up
up:
	docker compose up -d --build


## down: псевдоним для "docker compose down"
.PHONY: down
down:
	docker compose down


## logs: псевдоним "docker compose logs -f"
.PHONY: logs
logs:
	docker compose logs -f


## clear: очистка volume хранилищ
.PHONY: clear
clear:
	docker volume rm proxy_postgres_data proxy_broker_data proxy_prometheus_data proxy_grafana_data


## depends: установка зависимостей
.PHONY: depends
depends:
	go mod download
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	go install github.com/pressly/goose/v3/cmd/goose@v3.24.2
	go install go.uber.org/mock/mockgen@v0.5.2


## test: запуск тестов
.PHONY: test
test:
	go test ./... -v --race


## generate: псевдоним для "go generate ./.."
.PHONY: generate
generate:
	go generate ./...


## lint: запуск линтера
.PHONY: lint
lint:
	golangci-lint run


## migrations-up: применение всех миграций БД
.PHONY: migrations-up
migrations-up:
	goose --dir migrations/ postgres postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB} up


## migrations-down: откат всех миграций БД
.PHONY: migrations-down
migrations-down:
	goose --dir migrations/ postgres postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB} down-to 0
