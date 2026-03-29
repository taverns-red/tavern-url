# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tavern ./cmd/tavern/

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Run stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/tavern .
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/migrate.sh ./migrate.sh
COPY --from=builder /go/bin/goose /usr/local/bin/goose

RUN chmod +x /app/migrate.sh

EXPOSE 8080

ENTRYPOINT ["./tavern"]
