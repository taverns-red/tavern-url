.PHONY: dev build test lint clean migrate templ docker

# --- Development ---

## Run the dev server with hot reload
dev: templ
	go run ./cmd/tavern/

## Generate Templ files
templ:
	templ generate

## Run tests
test:
	go test ./... -v -count=1

## Run tests with coverage
test-cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Run linter
lint:
	golangci-lint run ./...

# --- Build ---

## Build the binary
build: templ
	CGO_ENABLED=0 go build -o bin/tavern ./cmd/tavern/

## Build Docker image
docker:
	docker build -t tavern-url .

# --- Database ---

## Run migrations up
migrate:
	goose -dir migrations postgres "$$DATABASE_URL" up

## Roll back one migration
migrate-down:
	goose -dir migrations postgres "$$DATABASE_URL" down

## Create a new migration
migrate-create:
	@read -p "Migration name: " name; \
	goose -dir migrations create $$name sql

# --- Cleanup ---

## Remove build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html tmp/
