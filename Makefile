.PHONY: dev test vet build migrate templ docker

# Development
dev:
	DATABASE_URL="postgres://tavern:tavern_dev@localhost:5432/tavern?sslmode=disable" go run ./cmd/tavern/

# Testing
test:
	go test ./... -count=1 -race

# Linting
vet:
	go vet ./...

# Build
build:
	go build -o tavern ./cmd/tavern/

# Database migrations
migrate:
	goose -dir migrations postgres "postgres://tavern:tavern_dev@localhost:5432/tavern?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://tavern:tavern_dev@localhost:5432/tavern?sslmode=disable" down

# Template generation
templ:
	templ generate

# Docker
docker-build:
	docker build -t tavern-url .

docker-run:
	docker run -p 8080:8080 \
		-e DATABASE_URL="postgres://tavern:tavern_dev@host.docker.internal:5432/tavern?sslmode=disable" \
		tavern-url
