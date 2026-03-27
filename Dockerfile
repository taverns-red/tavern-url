# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Install Templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Generate Templ files and build
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o /tavern ./cmd/tavern/

# Runtime stage
FROM gcr.io/distroless/static-debian12

COPY --from=builder /tavern /tavern
COPY --from=builder /app/static /static
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/tavern"]
