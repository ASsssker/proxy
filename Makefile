ifeq ("$(wildcard .env)","")
$(shell cp .env-example .env)
endif

include .env

## help: вывод информации о командах
.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


## depends: установка зависимостей
.PHONY: depends
depends:
	go mod download
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

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
