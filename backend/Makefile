.PHONY: dev build test lint migrate-up migrate-down seed clean docker-up docker-down

# Development
dev:
	go run cmd/server/main.go

build:
	go build -o bin/xexplay-api cmd/server/main.go

# Testing
test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

# Database Migrations
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

seed:
	go run scripts/seed.go

# Docker
docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

docker-build:
	docker build -f docker/api.Dockerfile -t xexplay-api .

# Clean
clean:
	rm -rf bin/
