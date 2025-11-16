# syntax=docker/dockerfile:1

# ------------ Builder Stage ------------ #
FROM golang:1.24 AS builder

# Install PostgreSQL client for pg_isready
RUN apt-get update && apt-get install -y postgresql-client

# Use repository root as build context inside the image so replace paths that point
# to other local modules (e.g. ../../pkg/kafka, ../../shared) resolve correctly.
WORKDIR /src

# Copy repository into the image. Using the repo root as build context (set in docker-compose)
# ensures the local modules referenced via `replace` in submodule go.mod files are available.
COPY . .

# Switch to the shipment-service module directory and download dependencies
WORKDIR /src/services/shipment-service
RUN go mod download

# Install goose for DB migration (will be copied to final image)
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/shipment-service ./cmd

# ------------ Final Stage (Slim Runtime) ------------ #
# Final runtime stage
FROM debian:bookworm-slim

# Install minimal PostgreSQL client
RUN apt-get update && apt-get install -y postgresql-client ca-certificates curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/shipment-service .
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy scripts from the builder stage (they were placed under /src during build)
# Using --from=builder avoids relying on host paths at final-stage COPY time.
COPY --from=builder /src/services/shipment-service/wait-for-postgres.sh ./
COPY --from=builder /src/services/shipment-service/entrypoint.sh ./
RUN chmod +x wait-for-postgres.sh entrypoint.sh

# Migrations folder (copy from builder workspace)
COPY --from=builder /src/services/shipment-service/db/migrations /migrations


# Expose port (optional if you're serving HTTP)
EXPOSE 8080

# Entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
