#!/bin/sh
#ENTRYPOINT.SH RUNS MIGRATIONS TO SET UP THE DATABASE

set -e # Exit on any error
# Analogy: The kitchen manager stops if setup fails

# Log database connection details for debugging
echo "🚀 Connecting to DB: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

# Print environment variables for debugging
echo "Environment vars:"
echo "  DB_USER=$DB_USER"
echo "  DB_PASSWORD=$DB_PASSWORD"
echo "  DB_NAME=$DB_NAME"
echo "  DB_HOST=$DB_HOST"
echo "  DB_PORT=$DB_PORT"

# Check if required environment variables are set
if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
  echo "❌ Missing one or more DB environment variables"
  env # Print all environment variables for debugging
  exit 1
fi

# Wait for Postgres to be ready
# Analogy: Ensure the pantry is ready before setting up the kitchen
./wait-for-postgres.sh "$DB_HOST" "$DB_PORT"

# Run database migrations
# Analogy: Build the pantry shelves using the blueprints (SQL migrations)
goose -dir /migrations postgres "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up

# Start the Shipment Service
# Analogy: Open the kitchen for business
./shipment-service