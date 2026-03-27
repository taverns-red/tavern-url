# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tavern ./cmd/tavern/

# Run stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/tavern .
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./tavern"]
