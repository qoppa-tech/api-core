ifneq (,$(wildcard ./.env.local))
	include .env.local
	export
else ifneq (,$(wildcard ./.env))
	include .env
	export
endif

DB_URL=postgresql://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable

all: build test

build:
	@echo "Building..."
	
	@go build -o main cmd/api/main.go

run:
	@go run cmd/api/main.go

docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

test:
	@echo "Testing..."
	@go test ./test/... -v

# itest:
# 	@echo "Running integration tests..."
# 	@go test ./internal/database -v

clean:
	@echo "Cleaning..."
	@rm -f main

watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi; \
	if ! command -v migrate > /dev/null; then \
		echo "golang-migrate is not installed. Run 'make install-migrate' first."; \
		exit 1; \
	fi; \
	migrate create -ext sql -dir migrations/schema -seq $(name)

migrate-up:
	@echo "Running migrations..."
	migrate -path migrations/schema -database "$(DB_URL)" -verbose up
migrate-down:
	@echo "Rolling back last migration..."
	migrate -path migrations/schema -database "$(DB_URL)" -verbose down 1

migrate-down-all:
	@echo "Rolling back all migrations..."
	migrate -path migrations/schema -database "$(DB_URL)" -verbose down -all
migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=001"; \
		exit 1; \
	fi; \
	migrate -path migrations/schema -database "$(DB_URL)" force $(version)

migrate-status:
	@if ! command -v migrate > /dev/null; then \
		echo "golang-migrate is not installed. Run 'make install-migrate' first."; \
		exit 1; \
	fi; \
	migrate -path migrations/schema -database "$(DB_URL)" version

queries_gen:
	@sqlc generate

.PHONY: all build run test clean watch docker-run docker-down itest
.PHONY: migrate-create migrate-up migrate-down migrate-down-all migrate-force migrate-status queries_gen
