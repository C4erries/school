.PHONY: all build run test lint compose-up compose-down compose-logs migrate-up migrate-down mock oapi

# Переменные
APP_NAME=api
BACKEND_DIR=backend

all: test build

build:
	@echo "==> Building backend..."
	cd $(BACKEND_DIR) && go build -o bin/$(APP_NAME) ./cmd/api

run:
	@echo "==> Running backend locally..."
	cd $(BACKEND_DIR) && go run ./cmd/api

test:
	@echo "==> Running tests..."
	cd $(BACKEND_DIR) && go test -v -race ./...

lint:
	@echo "==> Running linter..."
	cd $(BACKEND_DIR) && golangci-lint run ./...

compose-up:
	@echo "==> Starting Docker Compose environment..."
	docker compose up -d

compose-down:
	@echo "==> Stopping Docker Compose environment..."
	docker compose down

compose-logs:
	@echo "==> Following logs..."
	docker compose logs -f

migrate-up:
	@echo "==> Running database migrations up..."
	docker compose run --rm migrate

mock:
	@echo "==> Generating mocks via mockery..."
	mockery

oapi:
	@echo "==> Generating OpenAPI server interface and models..."
	cd $(BACKEND_DIR) && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=api/openapi/oapi-codegen.yaml api/openapi/api.yaml

