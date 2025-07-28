#!/bin/sh
set -e

echo "🚀 Connecting to DB: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

# Debugging variables
echo "Environment vars:"
echo "  DB_USER=$DB_USER"
echo "  DB_PASSWORD=$DB_PASSWORD"
echo "  DB_NAME=$DB_NAME"
echo "  DB_HOST=$DB_HOST"
echo "  DB_PORT=$DB_PORT"

# Check for required environment variables
if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ] || [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ]; then
  echo "❌ Missing one or more DB environment variables"
  env
  exit 1
fi

# Wait until PostgreSQL is ready
./wait-for-postgres.sh "$DB_HOST" "$DB_PORT"

# Run DB migrations
echo "⚙️ Running DB migrations..."
goose -dir /migrations postgres "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up

# Start the service
echo "✅ Starting shipment-service..."
exec ./shipment-service
