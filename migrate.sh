#!/bin/sh
# migrate.sh — Fly.io release command
# Applies goose migrations from /app/migrations using DATABASE_URL.

set -e

echo "Running database migrations..."
goose -dir /app/migrations postgres "$DATABASE_URL" up
echo "Migrations complete."
