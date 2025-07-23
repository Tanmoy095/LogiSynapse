#!/bin/sh
set -e

echo "🚀 Connecting to DB: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

# Print variables for debugging
echo "Environment vars:"
echo "  DB_USER=$DB_USER"
echo "  DB_PASSWORD=$DB_PASSWORD"
echo "  DB_NAME=$DB_NAME"
echo "  DB_HOST=$DB_HOST"
echo "  DB_PORT=$DB_PORT"

if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
  echo "❌ Missing one or more DB environment variables"
  env
  exit 1
fi

./wait-for-postgres.sh "$DB_HOST" "$DB_PORT"

goose -dir /migrations postgres "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up

./shipment-service
