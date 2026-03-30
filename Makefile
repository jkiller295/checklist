.PHONY: dev build up down logs setup tidy

# Generate go.sum then run locally (requires gcc for cgo)
tidy:
	go mod tidy

dev: tidy
	DB_PATH=./checklist.db go run ./cmd/main.go

build: tidy
	CGO_ENABLED=1 go build -o checklist ./cmd/main.go

# Docker
up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

setup:
	cp -n .env.example .env || true
	@echo "Edit .env to set HOUSEHOLD_PASSWORD, then run: make up"
