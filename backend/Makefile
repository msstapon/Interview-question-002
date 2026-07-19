.PHONY: help run build tidy test migrate-up migrate-down migrate-status migrate-new db-create keys

APP_NAME := interview-question-002

-include .env
export

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

run: ## Run the API server
	@go run ./cmd/api

build: ## Build the API binary into tmp/
	@go build -ldflags="-s -w" -o tmp/$(APP_NAME) ./cmd/api

tidy: ## go mod tidy
	@go mod tidy

test: ## Run unit tests
	@go test -race -count=1 ./...

db-create: ## Create the target database if it does not exist
	@go run ./cmd/migrate create-db

migrate-up: ## Apply migrations
	@go run ./cmd/migrate up

migrate-down: ## Rollback last migration (override: make migrate-down n=2)
	@go run ./cmd/migrate down $(or $(n),1)

migrate-status: ## Show migration status
	@go run ./cmd/migrate status

migrate-new: ## Create a new migration: make migrate-new name=add_foo
	@go run ./cmd/migrate new $(name)

keys: ## Generate the RS256 JWT keypair into ./secrets
	@go run ./cmd/keys
