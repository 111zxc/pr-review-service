#!/bin/sh
set -e

echo "Running migrations..."

go install github.com/pressly/goose/v3/cmd/goose@latest

goose -dir migrations postgres \
  "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE" up

echo "Migrations completed."
