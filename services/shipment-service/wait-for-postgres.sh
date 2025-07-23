#!/bin/sh
#IT CALLED BY ENTRYPOINT.SH .. IT ENSURE POSTGRES IS READY AND START SHIPMENT SERVICE
set -e # Exit on any error
# Analogy: The assistant stops if something goes wrong

host="$1" # Database host (e.g., "postgres")
port="$2" # Database port (e.g., "5432")
shift 2
cmd="$@" # Command to run after Postgres is ready (e.g., run migrations)

# Try 30 times to check if Postgres is ready
for i in $(seq 1 30); do
  if pg_isready -h "$host" -p "$port" -U "$DB_USER"; then
    echo "Postgres is up at $host:$port"
    break
  fi
  echo "Waiting for Postgres at $host:$port (attempt $i/30)..."
  sleep 3 # Wait 3 seconds before retrying
done

# Fail if Postgres isn't ready after 30 attempts
if ! pg_isready -h "$host" -p "$port" -U "$DB_USER"; then
  echo "Error: Postgres at $host:$port is not ready after 30 attempts"
  exit 1
fi

# Run the command (e.g., migrations and service)
echo "Postgres is ready - executing command"
exec $cmd