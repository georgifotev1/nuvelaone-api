.PHONY: run build test lint tidy migrate-up migrate-down docker-build docker-up docker-down docker-logs certbot swagger

-include .env
export

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Docker
docker-build:
	docker build -t nuvelaone-api .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-ps:
	docker-compose ps

# Production
docker-prod-build:
	docker-compose -f docker-compose.production.yaml build

docker-prod-up:
	docker-compose -f docker-compose.production.yaml up -d

docker-prod-down:
	docker-compose -f docker-compose.production.yaml down

docker-prod-logs:
	docker-compose -f docker-compose.production.yaml logs -f

docker-prod-restart:
	docker-compose -f docker-compose.production.yaml restart api

# SSL certificates (run once)
certbot:
	docker-compose -f docker-compose.production.yaml run --rm certbot

# Migrations
migrate-up:
	@source .env && goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	@source .env && goose -dir migrations postgres "$(DATABASE_URL)" down

# Database backup
db-backup:
	PGPASSWORD=$(DB_PASSWORD) pg_dump -h localhost -U $(DB_USER) $(DB_NAME) > backup_$$(date +%Y%m%d_%H%M%S).sql
